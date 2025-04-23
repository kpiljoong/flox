package filters

import (
	"os"
	"regexp"
)

type JSONFilter struct {
	DropFields   []string
	RenameFields map[string]string
	AddFields    map[string]string
}

var envVarPattern = regexp.MustCompile(`^\$\{(\w+)\}$`)

func NewJSONFilter(drop []string, rename, add map[string]string) *JSONFilter {
	return &JSONFilter{
		DropFields:   drop,
		RenameFields: rename,
		AddFields:    add,
	}
}

func (f *JSONFilter) Process(event map[string]interface{}) map[string]interface{} {
	// Drop
	for _, key := range f.DropFields {
		delete(event, key)
	}

	// Rename
	for oldKey, newKey := range f.RenameFields {
		if value, exists := event[oldKey]; exists {
			event[newKey] = value
			delete(event, oldKey)
		}
	}

	// Enrich
	for key, value := range f.AddFields {
		if matches := envVarPattern.FindStringSubmatch(value); matches != nil {
			envVal := os.Getenv(matches[1])
			event[key] = envVal
		} else {
			event[key] = value
		}
	}

	return event
}
