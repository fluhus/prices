package bouncer

// Handles state persistence for different runs.

import (
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"io"
	"os"
)

// Keeps state for continuing a previous run.
type stateType struct {
	Items []*Item
	ItemsMap map[string][]int
	ItemMetaMap    map[string][]*itemMetaId
}

// Current state.
var state *stateType

// Output files, for merging temp files into concrete ones.
var outFiles map[string]struct{}

// Loads persistence data from the output folder.
func initPersistence() {
	state = &stateType{}
	err := loadState(filepath.Join(outDir, "state"), state)
	if err != nil {
		// TODO(amit): Raise error only if not not-exist.
	}
	outFiles = map[string]struct{}{}
}

// Saves persistence data to the output folder.
func finalizePersistence() {
	err := saveState(filepath.Join(outDir, "state"), state)
	if err != nil {
		panic(err)
	}
	
	// Merge temp files to permanent files.
	for file := range outFiles {
		in, err := newFileReader(file + ".temp")
		if err != nil {
			panic(err)
		}
		out, err := newFileAppender(file)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(out, in)
		if err != nil {
			panic(err)
		}
		in.Close()
		out.Close()
		os.Remove(file + ".temp")
	}
}

func loadState(file string, a interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil { return err }
	return json.Unmarshal(data, a)
}

func saveState(file string, a interface{}) error {
	data, err := json.MarshalIndent(a, "", "\t")
	if err != nil { return err }
	return ioutil.WriteFile(file, data, 0644)
}

