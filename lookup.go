package geoiplookup

import (
	"bufio"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"
)

// Handler struct.
type Handler struct {
	Db   *bolt.DB
	Opts Options
}

// New returns a new handler.
func New(opt ...Option) *Handler {
	opts := NewOptions(opt...)

	db, err := bolt.Open(opts.DBPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &Handler{
		Db:   db,
		Opts: opts,
	}
}

// Close will close the db connection.
func (h *Handler) Close() {
	h.Db.Close()
}

// Lookup returns the country code given the ipv4/ipv6 address.
func (h *Handler) Lookup(ip string) string {
	countryCode := "NA"
	if !govalidator.IsIP(ip) {
		return countryCode
	}

	switch {
	case govalidator.IsIPv4(ip):
		return h.lookup(ip, IPv4)
	case govalidator.IsIPv6(ip):
		return h.lookup(ip, IPv6)
	}

	return countryCode
}

func (h *Handler) lookup(ip, ipVersion string) string {
	ipNet := net.ParseIP(ip)

	countryCode := "NA"
	var bucket string
	var byteCount int

	switch ipVersion {
	case IPv4:
		bucket = BoltBucketv4
		byteCount = IPv4ByteCount
	case IPv6:
		bucket = BoltBucketv6
		byteCount = IPv6ByteCount
	}

	for i := 0; i < byteCount; i++ {
		mask := net.CIDRMask(i, byteCount)
		network := ipNet.Mask(mask).String() + "/" + strconv.Itoa(i)

		if err := h.Db.View(func(tx *bolt.Tx) error {
			v := tx.Bucket([]byte(bucket)).Get([]byte(network))
			if v != nil {
				countryCode = string(v)
			}
			return nil
		}); err != nil {
			log.Printf("failed to get key %s: %v", network, err)
		}
	}

	return countryCode
}

// IterateDB iterates over a given bucket in DB.
func (h *Handler) IterateDB() {
	if err := h.Db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BoltBucketv6)).ForEach(func(k []byte, v []byte) error {
			log.Printf("key = %s , value = %s", string(k), string(v))
			return nil
		})
	}); err != nil {
		log.Printf("failed to iterate: %v", err)
	}
}

// PopulateData extracts the geoip data for each RIR and populates the database.
func (h *Handler) PopulateData() error {
	h.InitializeBuckets()

	fileNames := make([]string, 0)

	for _, rirURL := range GeoIPDataURLs {
		u, _ := url.Parse(rirURL)
		fileName := path.Base(u.EscapedPath())
		fileNames = append(fileNames, fileName)
	}

	// Download the RIR files with ip data
	if h.Opts.DownloadRIRFiles {
		if err := h.downloadRIRFiles(); err != nil {
			return errors.Wrap(err, "failed to download rir files")
		}
	}

	// Process the RIR files with ip data
	if err := h.processRIRFiles(fileNames); err != nil {
		return errors.Wrap(err, "failed to process rir files")
	}

	return nil
}

// downloadRIRFiles downloads the geoip data to files.
func (h *Handler) downloadRIRFiles() error {
	var g errgroup.Group

	for _, rirURL := range GeoIPDataURLs {
		currURL := rirURL
		u, _ := url.Parse(currURL)
		fileName := path.Base(u.EscapedPath())

		g.Go(func() error {
			log.Printf("downloading %s", currURL)
			rsp, err := h.Opts.HTTPClient.Get(currURL)
			if err != nil {
				return errors.Wrap(err, "failed to get geoip data")
			}
			if rsp.StatusCode != http.StatusOK {
				return errors.Errorf("failed to get geoip data with status %d", rsp.StatusCode)
			}
			defer rsp.Body.Close()

			log.Printf("saving data to file: %s", fileName)
			file, err := os.Create(fileName)
			if err != nil {
				return errors.Wrap(err, "failed to create local file")
			}
			defer file.Close()

			io.Copy(file, rsp.Body)

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// processRIRFiles processes the downloaded rir files and updates db.
func (h *Handler) processRIRFiles(fileNames []string) error {
	var g errgroup.Group

	for _, f := range fileNames {
		fileName := f

		g.Go(func() error {
			log.Printf("processing %s", fileName)

			file, err := os.Open(fileName)
			if err != nil {
				return errors.Wrapf(err, "failed to open file %s", fileName)
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				var country, ipVersion, ipAddress, mask string
				parts := strings.Split(scanner.Text(), "|")

				if len(parts) > 4 {
					country, ipVersion, ipAddress, mask = parts[1], parts[2], parts[3], parts[4]
				}

				if err := h.handleIP(ipAddress, country, mask, ipVersion); err != nil {
					continue
				}
			}

			if err := scanner.Err(); err != nil {
				return errors.Wrap(err, "error when reading ip data")
			}

			return nil
		})
	}

	return g.Wait()
}

func (h *Handler) handleIP(ip, country, mask, ipVersion string) error {
	var bucket string
	var byteCount int

	switch ipVersion {
	case IPv4:
		bucket = BoltBucketv4
		byteCount = IPv4ByteCount
	case IPv6:
		bucket = BoltBucketv6
		byteCount = IPv6ByteCount
	default:
		return errors.New("unrecognised ip version")
	}

	if country == "" {
		return errors.New("country not set")
	}

	count, err := strconv.Atoi(mask)
	if err != nil {
		return errors.Wrap(err, "failed to parse ip mask")
	}

	subnet := computeSubnet(ip, count, byteCount)
	if err := h.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Put([]byte(subnet), []byte(country))
	}); err != nil {
		return errors.Wrap(err, "failed to update k/v in db")
	}

	return nil
}

// InitializeBuckets handler initializes the buckets in DB.
func (h *Handler) InitializeBuckets() error {
	// Create IPv4 bucket in db.
	if err := h.createBucket(BoltBucketv4); err != nil {
		return errors.Wrap(err, "failed to create ipv4 bucket in db")
	}

	// Create IPv6 bucket in db.
	if err := h.createBucket(BoltBucketv6); err != nil {
		return errors.Wrap(err, "failed to create ipv6 bucket in db")
	}

	return nil
}

// CreateBucket creates the given bucket in boltdb if it doesn't exist.
func (h *Handler) createBucket(bucket string) error {
	if err := h.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		return err
	}); err != nil && err != bolt.ErrBucketExists {
		return errors.Wrap(err, "failed to update k/v in db")
	}
	return nil
}

func computeSubnet(ipstart string, ipcount, byteCount int) string {
	mask := byteCount - int(math.Log2(float64(ipcount)))
	return ipstart + "/" + strconv.Itoa(mask)
}

var (
	// GeoIPDataURLs is the string slice containing URL for each RIR.
	GeoIPDataURLs = []string{Arin, RipeNcc, Apnic, Afrinic, Lacnic}
)

// Constants used in geoiplookup.
const (
	// BoltBucketv4 containing ipv4 data
	BoltBucketv4 = "ipv4"
	// BoltBucketv6 containing ipv6 data
	BoltBucketv6 = "ipv6"

	// IPv4 represents ipv4 address
	IPv4 = "ipv4"
	// IPv6 represents ipv6 address
	IPv6 = "ipv6"

	// IPv4ByteCount is the ipv4 byte count
	IPv4ByteCount = 32
	// IPv6ByteCount is the ipv6 byte count
	IPv6ByteCount = 128

	// URLs for each RIR containing geoip data.
	Arin    = "https://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest"
	RipeNcc = "https://ftp.ripe.net/ripe/stats/delegated-ripencc-extended-latest"
	Apnic   = "https://ftp.apnic.net/stats/apnic/delegated-apnic-extended-latest"
	Afrinic = "https://ftp.apnic.net/stats/afrinic/delegated-afrinic-extended-latest"
	Lacnic  = "https://ftp.apnic.net/stats/lacnic/delegated-lacnic-extended-latest"
)
