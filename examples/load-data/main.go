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
		// goiplookup.WithDownloadRIRFiles(),
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
