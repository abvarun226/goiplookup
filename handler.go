package goiplookup

import (
	"log"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
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

var (
	// GeoIPDataURLs is the string slice containing URL for each RIR.
	GeoIPDataURLs = []string{Arin, RipeNcc, Apnic, Afrinic, Lacnic}
)

// Constants used in goiplookup.
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
