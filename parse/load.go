package main

// An interface for loading data from various file types.

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

// loadLock synchronizes calls to load(), to prevent hard-drive from being
// accessed in parallel.
var loadLock sync.Mutex

// Loads data from a file, and decompresses if it is a gzip or a zip.
func load(file string) ([]byte, error) {
	loadLock.Lock()
	defer loadLock.Unlock()

	switch {
	// Gzip.
	case strings.HasSuffix(file, ".gz"):
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b := bufio.NewReader(f)
		z, err := gzip.NewReader(b)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(z)
		if err != nil {
			return nil, err
		}
		return data, nil

	// Zip.
	case strings.HasSuffix(file, ".zip"):
		z, err := zip.OpenReader(file)
		if err != nil {
			return nil, err
		}
		if len(z.File) != 1 {
			return nil, fmt.Errorf("Zip should have 1 file, but has %d "+
				"instead.", len(z.File))
		}
		f, err := z.File[0].Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return data, nil

	// Plain text.
	default:
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b := bufio.NewReader(f)
		data, err := ioutil.ReadAll(b)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}
