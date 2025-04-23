package file

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const stateFile = ".flox.state"

var stateLock sync.Mutex

type OffsetState map[string]int64

func loadState() OffsetState {
	stateLock.Lock()
	defer stateLock.Unlock()
	return loadStateUnlocked()
}

func loadStateUnlocked() OffsetState {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return make(OffsetState)
	}

	var state OffsetState
	if err := json.Unmarshal(data, &state); err != nil {
		return make(OffsetState)
	}

	return state
}

func saveOffset(path string, offset int64) {
	stateLock.Lock()
	defer stateLock.Unlock()

	state := loadStateUnlocked()
	state[path] = offset

	data, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		log.Printf("Failed to marshal state: %v", err)
		return
	}
	if err := os.WriteFile(stateFile, data, 0o644); err != nil {
		log.Printf("Failed to write state: %v", err)
	}
}
