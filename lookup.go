package goiplookup

import (
	"fmt"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// Lookup returns the country code given the ipv4/ipv6 address.
func (h *Handler) Lookup(ip string) (string, error) {
	countryCode := "unknown"
	if !govalidator.IsIP(ip) {
		return countryCode, errors.New("not a valid ip")
	}

	switch {
	case govalidator.IsIPv4(ip):
		return h.lookup(ip, IPv4)
	case govalidator.IsIPv6(ip):
		return h.lookup(ip, IPv6)
	}

	return countryCode, errors.New("unknown ip address")
}

func (h *Handler) lookup(ip, ipVersion string) (string, error) {
	ipNet := net.ParseIP(ip)

	countryCode := "unknown"
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

	finalErr := fmt.Errorf("ip not found")
	for i := 0; i < byteCount; i++ {
		mask := net.CIDRMask(i, byteCount)
		network := ipNet.Mask(mask).String() + "/" + strconv.Itoa(i)

		h.Db.View(func(tx *bolt.Tx) error {
			v := tx.Bucket([]byte(bucket)).Get([]byte(network))
			if v != nil {
				countryCode = string(v)
				finalErr = nil
			}
			return nil
		})
	}

	return countryCode, finalErr
}

// IterateDB iterates over a given bucket in DB.
func (h *Handler) IterateDB(bucket string) error {
	if err := h.Db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).ForEach(func(k []byte, v []byte) error {
			fmt.Printf("key = %s , value = %s\n", string(k), string(v))
			return nil
		})
	}); err != nil {
		return errors.Wrap(err, "failed to iterate")
	}

	return nil
}
