package main

import (
	"net/http"
	"log"
	"fmt"
	"time"
	"encoding/json"
	"sync"

	"github.com/abvarun226/goiplookup"
	"github.com/asaskevich/govalidator"
)

func main() {
	mux := http.NewServeMux()
	ipLookup := goiplookup.New(
		goiplookup.WithDBPath(DBPath),
		goiplookup.WithClient(&http.Client{Timeout: 60 * time.Minute}),
		goiplookup.WithDownloadRIRFiles(),
	)
	defer ipLookup.Close()

	go func() {
		for {
			time.Sleep(1440 * time.Minute)
			log.Printf("start populating geo ip data")
			if errPop := ipLookup.PopulateData(); errPop != nil {
				log.Printf("error populating geoip data: %v", errPop)
			}
			log.Printf("completed populating geo ip data")
		}
	}()

	h := New(ipLookup)
	mux.HandleFunc("/iplookup", h.LookupEndpoint)
	http.ListenAndServe(":8085", mux)
}

// LookupEndpoint is the http endpoint for ip address lookup.
func (h *Handler) LookupEndpoint(w http.ResponseWriter, r *http.Request) {
	ips, ok := r.URL.Query()["ip"]	
	if !ok || len(ips) == 0 {
		log.Printf("user did not provide ip parameter")
		http.Error(w, "user did not provide ip parameter", http.StatusBadRequest)
		return
	}

	result := make(chan lookupResult, len(ips))

	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(wg *sync.WaitGroup, ip string, result chan lookupResult) {
			defer wg.Done()
			res := lookupResult{IP: ip}
			var err error

			res.Country , err = h.IPLookup.Lookup(ip)
			if err != nil {
				log.Printf("error when looking up ip address %s: %v", ip, err)
				res.Err = fmt.Errorf("failed to lookup ip address")
			}
		
			switch {
			case govalidator.IsIPv4(ip):
				res.Version = "ipv4"
			case govalidator.IsIPv6(ip):
				res.Version = "ipv6"
			default:
				res.Version = "unknown"
			}

			result <- res
		}(&wg, ip, result)
	}

	wg.Wait()
	close(result)

	rsp := make([]lookupResult, 0)
	for m := range result {
		rsp = append(rsp, m)
	}

	jsonRsp, err := json.Marshal(rsp)
	if err != nil {
		log.Printf("error when marshaling response: %v", err)
		http.Error(w, "failed to render response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonRsp)
}

type lookupResult struct {
	Country string `json:"country"`
	IP string `json:"ip"`
	Version string `json:"version"`
	Err error `json:"error,omitempty"`
}

const (
	// DBPath is the location of bolt db file.
	DBPath = "../my.db"
)
