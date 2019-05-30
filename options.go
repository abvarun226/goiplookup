package goiplookup

import (
	"net/http"
	"time"
)

// Options struct.
type Options struct {
	DBPath           string
	HTTPClient       *http.Client
	DownloadRIRFiles bool
	RIRFilesDir      string
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

// WithRIRFilesDir sets the directory to download RIR files to.
func WithRIRFilesDir(dirpath string) Option {
	return func(o *Options) {
		o.RIRFilesDir = dirpath
	}
}

// NewOptions returns a new Options object.
func NewOptions(options ...Option) Options {
	opts := Options{
		DBPath:      DefaultDBPath,
		HTTPClient:  &http.Client{Timeout: DefaultTimeout},
		RIRFilesDir: DefaultFileDir,
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
