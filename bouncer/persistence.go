package bouncer

// Handles state persistence for different runs.

import (
	"path/filepath"
	"os"

	"github.com/fluhus/gostuff/gobz"
)

var (
	persistenceData map[string]interface{} // Persists state across different runs.
)

// Loads persistence data from the output folder.
func initPersistence() {
	persistenceData = map[string]interface{}{}
	err := gobz.Load(filepath.Join(outDir, "state.gobz"), &persistenceData)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

// Saves persistence data to the output folder.
func finalizePersistence() {
	err := gobz.Save(filepath.Join(outDir, "state.gobz"), persistenceData)
	if err != nil {
		panic(err)
	}
}

