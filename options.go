package geoiplookup

import (
	"net/http"
	"time"
)

// Options struct.
type Options struct {
	DBPath           string
	HTTPClient       *http.Client
	DownloadRIRFiles bool
}

// Option type.
type Option func(*Options)

// WithDBPath sets db file location.
func WithDBPath(dbPath string) Option {
	return func(o *Options) {
		o.DBPath = dbPath
	}
}

// WithDownloadRIRFiles ensures that the rir files are
// downloaded to populate new geoip data.
func WithDownloadRIRFiles() Option {
	return func(o *Options) {
		o.DownloadRIRFiles = true
	}
}

// WithClient sets HTTP client.
func WithClient(client *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

// NewOptions returns a new Options object.
func NewOptions(options ...Option) Options {
	opts := Options{
		DBPath:     DefaultDBPath,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}

	for _, o := range options {
		o(&opts)
	}

	return opts
}

const (
	// DefaultDBPath is the default location of bolt db file.
	DefaultDBPath = "/tmp/geoip.db"

	// DefaultTimeout is the default http client timeout.
	DefaultTimeout = 180 * time.Second
)
