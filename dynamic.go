package goenum

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EnumDefinition represents the structure for loading enum data
type EnumDefinition struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	Aliases     []string    `json:"aliases,omitempty"`
}

// DynamicEnumLoader provides functionality to load enums from various sources
type DynamicEnumLoader struct {
	enumSet *EnumSet[Enum]
}

// NewDynamicEnumLoader creates a new DynamicEnumLoader instance
func NewDynamicEnumLoader() *DynamicEnumLoader {
	return &DynamicEnumLoader{
		enumSet: NewEnumSet[Enum](),
	}
}

// LoadFromJSON loads enum definitions from a JSON file
func (l *DynamicEnumLoader) LoadFromJSON(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return l.LoadFromReader(file)
}

// LoadFromReader loads enum definitions from an io.Reader
func (l *DynamicEnumLoader) LoadFromReader(reader io.Reader) error {
	var definitions []EnumDefinition
	if err := json.NewDecoder(reader).Decode(&definitions); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	for _, def := range definitions {
		// Convert float64 to int if necessary
		if f, ok := def.Value.(float64); ok {
			def.Value = int(f)
		}
		enum := &EnumBase{
			name:        def.Name,
			value:       def.Value,
			description: def.Description,
			aliases:     def.Aliases,
			jsonConfig:  DefaultJSONConfig(),
		}
		l.enumSet.Register(enum)
	}

	return nil
}

// LoadFromDirectory loads all JSON files from a directory
func (l *DynamicEnumLoader) LoadFromDirectory(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no JSON files found in directory: %s", dir)
	}

	for _, file := range files {
		if err := l.LoadFromJSON(file); err != nil {
			return fmt.Errorf("failed to load file %s: %w", file, err)
		}
	}

	return nil
}

// GetEnumSet returns the loaded enum set
func (l *DynamicEnumLoader) GetEnumSet() *EnumSet[Enum] {
	return l.enumSet
}

// LoadFromMap loads enum definitions from a map
func (l *DynamicEnumLoader) LoadFromMap(definitions map[string]EnumDefinition) error {
	for _, def := range definitions {
		enum := &EnumBase{
			name:        def.Name,
			value:       def.Value,
			description: def.Description,
			aliases:     def.Aliases,
			jsonConfig:  DefaultJSONConfig(),
		}
		l.enumSet.Register(enum)
	}
	return nil
}

// LoadFromSlice loads enum definitions from a slice
func (l *DynamicEnumLoader) LoadFromSlice(definitions []EnumDefinition) error {
	for _, def := range definitions {
		enum := &EnumBase{
			name:        def.Name,
			value:       def.Value,
			description: def.Description,
			aliases:     def.Aliases,
			jsonConfig:  DefaultJSONConfig(),
		}
		l.enumSet.Register(enum)
	}
	return nil
}

// ExportToJSON exports the current enum set to a JSON file
func (l *DynamicEnumLoader) ExportToJSON(filename string) error {
	definitions := make([]EnumDefinition, 0)
	for _, enum := range l.enumSet.Values() {
		definitions = append(definitions, EnumDefinition{
			Name:        enum.String(),
			Value:       enum.Value(),
			Description: enum.Description(),
			Aliases:     enum.Aliases(),
		})
	}

	data, err := json.MarshalIndent(definitions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal enums: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
