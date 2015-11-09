// Simple serialization functions.
// A gobz is simply a gzipped gob. Functions here allow quick
// reading and writing of gobzs.
package gobz

import (
	"fmt"
	"os"
	"io"
	"bufio"
	"compress/gzip"
	"encoding/gob"
)

// Writes a value to the given stream.
func Write(w io.Writer, obj interface{}) error {
	// Open zip stream.
	z := gzip.NewWriter(w)
	defer z.Close()
	
	// Write data.
	err := gob.NewEncoder(z).Encode(obj)
	if err != nil {
		return fmt.Errorf("Could not encode object: %v", err)
	} else {
		return nil
	}
}

// Reads a value from the given stream.
func Read(r io.Reader, obj interface{}) error {
	// Open zip stream.
	z, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("Could not read gzip: %v", err)
	}
	
	// Read data.
	err = gob.NewDecoder(z).Decode(obj)
	if err != nil {
		return fmt.Errorf("Could not decode object: %v", err)
	}
	
	err = z.Close()
	if err != nil {
		return fmt.Errorf("Could not read gzip: %v", err)
	}
	
	return nil
}

// Writes a value to the given file.
func Save(file string, obj interface{}) error {
	// Open file.
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("Could not open file: %v", err)
	}
	defer f.Close()
	
	b := bufio.NewWriter(f)
	defer b.Flush()
	
	return Write(b, obj)
}

// Reads a value from the given file.
func Load(file string, obj interface{}) error {
	// Open file.
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Could not open file: %v", err)
	}
	defer f.Close()
	
	b := bufio.NewReader(f)
	
	return Read(b, obj)
}

