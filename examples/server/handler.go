package main

import (
	"github.com/abvarun226/goiplookup"
)

// Handler struct
type Handler struct {
	IPLookup *goiplookup.Handler
}

// New returns a new initialized handler.
func New(ipLookup *goiplookup.Handler) *Handler {
	return &Handler{IPLookup: ipLookup}
}