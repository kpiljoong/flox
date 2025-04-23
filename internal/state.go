package internal

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const stateFile = ".flox.state"

var stateLock sync.Mutex

type OffsetState map[string]int64

func LoadState() OffsetState {
	stateLock.Lock()
	defer stateLock.Unlock()
	return LoadStateUnlocked()
}

func LoadStateUnlocked() OffsetState {
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

func SaveOffset(path string, offset int64) {
	stateLock.Lock()
	defer stateLock.Unlock()

	state := LoadStateUnlocked()
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
