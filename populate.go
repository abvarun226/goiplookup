package goiplookup

import (
	"bufio"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"
)

// PopulateData extracts the geoip data for each RIR and populates the database.
func (h *Handler) PopulateData() error {
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
	os.Mkdir(h.Opts.RIRFilesDir, 0755)

	for _, rirURL := range GeoIPDataURLs {
		currURL := rirURL
		u, _ := url.Parse(currURL)
		fileName := path.Base(u.EscapedPath())

		g.Go(func() error {
			log.Printf("downloading %s", currURL)

			// Get RIR files.
			rsp, err := h.Opts.HTTPClient.Get(currURL)
			if err != nil {
				return errors.Wrap(err, "failed to get geoip data")
			}
			if rsp.StatusCode != http.StatusOK {
				return errors.Errorf("failed to get geoip data with status %d", rsp.StatusCode)
			}
			defer rsp.Body.Close()

			log.Printf("saving data to file: %s", fileName)
			path := h.Opts.RIRFilesDir + "/" + fileName
			file, err := os.Create(path)
			if err != nil {
				return errors.Wrap(err, "failed to create local file")
			}
			defer file.Close()

			io.Copy(file, rsp.Body)

			return nil
		})
	}
	return g.Wait()
}

// processRIRFiles processes the downloaded rir files and updates db.
func (h *Handler) processRIRFiles(fileNames []string) error {
	var g errgroup.Group
	os.Mkdir(h.Opts.RIRFilesDir, 0755)

	for _, f := range fileNames {
		fileName := f

		g.Go(func() error {
			log.Printf("processing %s", fileName)

			path := h.Opts.RIRFilesDir + "/" + fileName
			file, err := os.Open(path)
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

func computeSubnet(ipstart string, ipcount, byteCount int) string {
	mask := byteCount - int(math.Log2(float64(ipcount)))
	return ipstart + "/" + strconv.Itoa(mask)
}
