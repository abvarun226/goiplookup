# Geo IP Lookup

[![Go Report Card](https://goreportcard.com/badge/github.com/abvarun226/geoiplookup)](https://goreportcard.com/report/github.com/abvarun226/geoiplookup)

This library maps an IP address to a country

## Description
It basically uses the Regional Internet Registries (APNIC, ARIN, LACNIC, RIPE NCC and AFRINIC) to populate a database with the IP address blocks allocated to a country. It then uses this database, which is refereshed on a daily basis, to lookup country to which an IP address belongs to.

## Examples
Examples in `examples/` directory.

### How to load data into database?
```
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/abvarun226/geoiplookup"
)

func main() {
	h := geoiplookup.New(
		geoiplookup.WithDBPath(DBPath),
		geoiplookup.WithClient(&http.Client{Timeout: 60 * time.Minute}),
		geoiplookup.WithDownloadRIRFiles(),
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

	"github.com/abvarun226/geoiplookup"
)

func main() {
	flag.Parse()

	if flag.NArg() <= 0 {
		fmt.Printf("please provide at least one ip address as an argument")
		os.Exit(1)
	}

	ips := flag.Args()

	h := geoiplookup.New(
		geoiplookup.WithDBPath(DBPath),
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