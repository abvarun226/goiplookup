# Golang IP Lookup Library

[![GoDoc](https://godoc.org/github.com/abvarun226/goiplookup?status.svg)](https://godoc.org/github.com/abvarun226/goiplookup)
[![Go Report Card](https://goreportcard.com/badge/github.com/abvarun226/goiplookup)](https://goreportcard.com/report/github.com/abvarun226/goiplookup)
[![Build Status](https://travis-ci.org/abvarun226/goiplookup.svg?branch=master)](https://travis-ci.org/abvarun226/goiplookup)
[![codecov](https://codecov.io/gh/abvarun226/goiplookup/branch/master/graph/badge.svg)](https://codecov.io/gh/abvarun226/goiplookup)

This library maps an IP address to a country without depending on external geo databases like Maxmind.

## Description
The library uses the Regional Internet Registries (APNIC, ARIN, LACNIC, RIPE NCC and AFRINIC) to populate a database with the IP address blocks allocated to a country. It then uses this database, which is refreshed on a daily basis, to lookup country to which an IP address belongs to.

## Examples
Examples in `examples/` directory.

`examples` directory also has an example of an HTTP server that uses this library.

### How to load data into database?
```
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/abvarun226/goiplookup"
)

func main() {
	h := goiplookup.New(
		goiplookup.WithDBPath(DBPath),
		goiplookup.WithClient(&http.Client{Timeout: 60 * time.Minute}),
		goiplookup.WithDownloadRIRFiles(),
	)
	defer h.Close()

	if errPop := h.PopulateData(); errPop != nil {
		log.Printf("error populating geoip data: %v", errPop)
	}
}

const (
	// DBPath is the location of bolt db file.
	DBPath = "../my.db"
)
```

### How to look up geo location?
```
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/abvarun226/goiplookup"
)

func main() {
	flag.Parse()

	if flag.NArg() <= 0 {
		fmt.Printf("please provide at least one ip address as an argument")
		os.Exit(1)
	}

	ips := flag.Args()

	h := goiplookup.New(
		goiplookup.WithDBPath(DBPath),
	)
	defer h.Close()

	for _, ip := range ips {
		country, err := h.Lookup(ip)
		if err != nil {
			fmt.Printf("error when looking up ip address: %v", err)
		}
		fmt.Printf("%15s : %s\n", ip, country)
	}
}

const (
	// DBPath is the location of bolt db file.
	DBPath = "../my.db"
)
```