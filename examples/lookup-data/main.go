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
		fmt.Printf("%15s : %s\n", ip, h.Lookup(ip))
	}
}

const (
	// DBPath is the location of bolt db file.
	DBPath = "../my.db"
)
