package goiplookup_test

import (
	"testing"
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/abvarun226/goiplookup"
)

func TestHandler_Lookup_IPv4(t *testing.T) {
	h, tmpfile := Setup()
	defer func() {
		h.Close()
		os.Remove(tmpfile)
	}()

	network := "49.204.0.0/14"
	ip := "49.206.13.16"
	expectedCountry := "IN"

	if err := h.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(goiplookup.BoltBucketv4)).
		Put([]byte(network), []byte(expectedCountry))
	}); err != nil {
		panic(err)
	}

	country, err := h.Lookup(ip)
	Convey("When calling handler.Lookup", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
		Convey(fmt.Sprintf("country ShouldEqual `%v`", expectedCountry), func() {
			So(country, ShouldEqual, expectedCountry)
		})
	})
}

func TestHandler_Lookup_IPv6(t *testing.T) {
	h,tmpfile := Setup()
	defer func() {
		h.Close()
		os.Remove(tmpfile)
	}()

	network := "2001:4c0::/123"
	ip := "2001:4c0:0:0:0:0:0:0"
	expectedCountry := "CA"

	if err := h.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(goiplookup.BoltBucketv6)).
		Put([]byte(network), []byte(expectedCountry))
	}); err != nil {
		panic(err)
	}

	country, err := h.Lookup(ip)
	Convey("When calling handler.Lookup", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
		Convey(fmt.Sprintf("country ShouldEqual `%v`", expectedCountry), func() {
			So(country, ShouldEqual, expectedCountry)
		})
	})
}

func TestHandler_Lookup_NotFound(t *testing.T) {
	h,tmpfile := Setup()
	defer func() {
		h.Close()
		os.Remove(tmpfile)
	}()

	ip := "49.206.13.16"
	expected := "unknown"

	country, err := h.Lookup(ip)
	Convey("When calling handler.Lookup", t, func() {
		Convey("err ShouldNotBeNil", func() {
			So(err, ShouldNotBeNil)
		})
		Convey(fmt.Sprintf("country ShouldEqual `%v`", expected), func() {
			So(country, ShouldEqual, expected)
		})
	})
}
