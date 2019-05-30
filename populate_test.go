package goiplookup_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/abvarun226/goiplookup"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandler_PopulateData(t *testing.T) {
	h, tmpfile := Setup()
	defer func() {
		h.Close()
		os.Remove(tmpfile)
		os.RemoveAll(goiplookup.DefaultFileDir)
	}()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `arin|US|ipv4|49.128.0.0|8388608|20171220|allocated|4a8a91b5b89d3f900098ebf73ca0b118
		arin|US|ipv6|2001:400::|32|19990803|allocated|04f048163e37eef48d891498545eefc0`
		w.Write([]byte(response))
	}))
	defer func() { testServer.Close() }()

	h.Opts.HTTPClient = testServer.Client()
	log.Print(testServer.URL)
	goiplookup.GeoIPDataURLs = []string{testServer.URL + "/arin-latest"}

	err := h.PopulateData()
	Convey("When calling handler.PopulateData", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
	})

	ip := "49.128.0.0"
	expected := "US"
	country, err := h.Lookup(ip)
	Convey("When calling handler.Lookup", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
		Convey(fmt.Sprintf("country ShouldEqual `%v`", expected), func() {
			So(country, ShouldEqual, expected)
		})
	})

	ip = "2001:400:0:0:0:0:0:0"
	expected = "US"
	country, err = h.Lookup(ip)
	Convey("When calling handler.Lookup", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
		Convey(fmt.Sprintf("country ShouldEqual `%v`", expected), func() {
			So(country, ShouldEqual, expected)
		})
	})
}
