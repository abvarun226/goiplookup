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
