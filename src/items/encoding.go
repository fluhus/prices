package main

import (
	"fmt"
	"regexp"
	"os/exec"
	"io/ioutil"
)

// Converts the given XML to utf-8, according to its 'encoding' field. Uses
// the shell command 'iconv' for the convertion. Also converts the field value
// to utf-8.
func toUtf8(text []byte) ([]byte, error) {
	match := regexp.MustCompile("encoding=\"(.*?)\"").FindSubmatch(text)
	
	// No encoding is assumed to be utf-8.
	if match == nil {
		return text, nil
	}
	
	// Handle encoding.
	enc := string(match[1])
	
	if enc == "" {
		return nil, fmt.Errorf("Encoding field is empty.")
	}
	
	if enc == "utf-8" || enc == "UTF-8" {
		return text, nil
	}
	
	// Convert to utf-8.
	cmd := exec.Command("iconv", "-f", enc, "-t", "utf-8")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	
	// Push input to stdin.
	go func() {
		in.Write(text)
		in.Close()
	}()

	// Read from stdout.
	newText, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}
	
	// End command.
	err = cmd.Wait()
	
	if err != nil {
		return nil, err
	}
	
	// Replace encoding field with utf-8.
	newText = regexp.MustCompile("encoding=\".*?\"").ReplaceAll(newText,
			[]byte("encoding=\"utf-8\""))
	
	return newText, nil
}

