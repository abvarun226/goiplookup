package goiplookup_test

import (
	"io/ioutil"
	"os"

	"github.com/abvarun226/goiplookup"
)

func Setup() (*goiplookup.Handler, string) {
	tmpfile := tempfile()

	h := goiplookup.New(
		goiplookup.WithDBPath(tmpfile),
		goiplookup.WithDownloadRIRFiles(),
		goiplookup.WithRIRFilesDir(goiplookup.DefaultFileDir),
	)

	if err := h.InitializeBuckets(); err != nil {
		panic(err)
	}

	return h, tmpfile
}

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
