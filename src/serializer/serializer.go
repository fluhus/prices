// Handles serialization and deserialization of parsed data.
package serializer

import (
	"bytes"
	"encoding/gob"
	"compress/gzip"
	"io/ioutil"
)

// Writes the given maps to the given file. The maps can be retreived later
// using a Deserializer.
func Serialize(file string, data []map[string]string) error {
	// Open encoding stream. Using buffer instead of file to write in one batch,
	// for better error handling.
	b := bytes.NewBuffer(nil)
	z := gzip.NewWriter(b)
	g := gob.NewEncoder(z)
	
	// Encode data.
	// Encoding each individually instead of the entire slice, to enble
	// streaming when deserializing.
	for _, datum := range data {
		g.Encode(datum)
	}
	
	// Write to file.
	z.Close()
	return ioutil.WriteFile(file, b.Bytes(), 0666)
}

// Reads parsed data from files. Reads each item separately.
type Deserializer struct {
	decoder *gob.Decoder
	err error
}

// Returns a new deserializer that reads from the given file. Error should be
// checked after creating the deserializer, before calling Next().
func NewDeserializer(file string) (d *Deserializer) {
	d = &Deserializer{}
	var err error
	
	b, err := ioutil.ReadFile(file)
	if err != nil {
		d.err = err
		return
	}
	z, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		d.err = err
		return
	}
	d.decoder = gob.NewDecoder(z)
	
	return
}

// Reads the next data item. Returns nil iff an error occurs, or if a previous
// error already exists.
func (d *Deserializer) Next() map[string]string {
	// If an error had already occurred, go no further.
	if d.err != nil {
		return nil
	}
	
	// Deserialize.
	var result map[string]string
	d.err = d.decoder.Decode(&result)
	if d.err != nil {
		return nil
	}
	
	return result
}

// Returns the last encountered error.
func (d *Deserializer) Err() error {
	return d.err
}








