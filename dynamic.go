package goenum

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

// DuplicateHandling defines how to handle duplicate enums during loading
type DuplicateHandling int

const (
	// DuplicateError will return an error when duplicates are found
	DuplicateError DuplicateHandling = iota
	// DuplicateSkip will skip duplicate entries
	DuplicateSkip
	// DuplicateOverride will override existing entries with new ones
	DuplicateOverride
)

// ValidationOptions defines options for enum validation
type ValidationOptions struct {
	// DuplicateHandling specifies how to handle duplicate enums
	DuplicateHandling DuplicateHandling
	// ValueType specifies the expected type for enum values (e.g., reflect.TypeOf(0) for int)
	ValueType reflect.Type
	// AllowEmptyNames allows enums with empty names
	AllowEmptyNames bool
	// AllowEmptyValues allows enums with nil values
	AllowEmptyValues bool
}

// DefaultValidationOptions returns the default validation options
func DefaultValidationOptions() *ValidationOptions {
	return &ValidationOptions{
		DuplicateHandling: DuplicateError,
		ValueType:         nil, // No type restriction by default
		AllowEmptyNames:   false,
		AllowEmptyValues:  false,
	}
}

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
	options *ValidationOptions
}

// NewDynamicEnumLoader creates a new DynamicEnumLoader instance
func NewDynamicEnumLoader(options *ValidationOptions) *DynamicEnumLoader {
	if options == nil {
		options = DefaultValidationOptions()
	}
	return &DynamicEnumLoader{
		enumSet: NewEnumSet[Enum](),
		options: options,
	}
}

// validateEnumDefinition validates an enum definition according to the options
func (l *DynamicEnumLoader) validateEnumDefinition(def EnumDefinition) error {
	// Check for empty name
	if !l.options.AllowEmptyNames && def.Name == "" {
		return fmt.Errorf("enum name cannot be empty")
	}

	// Check for empty value
	if !l.options.AllowEmptyValues && def.Value == nil {
		return fmt.Errorf("enum value cannot be nil")
	}

	// Check value type if specified
	if l.options.ValueType != nil && def.Value != nil {
		valueType := reflect.TypeOf(def.Value)
		if !valueType.AssignableTo(l.options.ValueType) {
			return fmt.Errorf("enum value type %v is not assignable to expected type %v",
				valueType, l.options.ValueType)
		}
	}

	return nil
}

// add validates one definition, applies the configured duplicate handling and
// registers it. It is the single path shared by LoadFromReader, LoadFromMap
// and LoadFromSlice, so all three agree on validation, duplicate detection
// (by name and by value) and error reporting.
func (l *DynamicEnumLoader) add(def EnumDefinition) error {
	if err := l.validateEnumDefinition(def); err != nil {
		return fmt.Errorf("invalid enum definition: %w", err)
	}
	_, dupName := l.enumSet.GetByName(def.Name)
	_, dupValue := l.enumSet.GetByValue(def.Value)
	if dupName || dupValue {
		switch l.options.DuplicateHandling {
		case DuplicateSkip:
			return nil
		case DuplicateOverride:
			l.overrideRemove(def.Name, def.Value)
		default: // DuplicateError
			return fmt.Errorf("duplicate enum found: name=%s, value=%v", def.Name, def.Value)
		}
	}
	l.enumSet.Register(&EnumBase{
		name:        def.Name,
		value:       def.Value,
		description: def.Description,
		aliases:     def.Aliases,
		jsonConfig:  DefaultJSONConfig(),
	})
	return nil
}

// overrideRemove drops any registered enum sharing the given name or value so
// the replacement can be registered without a duplicate panic.
// ponytail: O(n) rebuild per override; fine for load-time use, revisit if
// overriding into large sets in a hot path.
func (l *DynamicEnumLoader) overrideRemove(name string, value interface{}) {
	newSet := NewEnumSet[Enum]()
	for _, enum := range l.enumSet.Values() {
		if enum.String() == name || enum.Value() == value {
			continue
		}
		newSet.Register(enum)
	}
	l.enumSet = newSet
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
		// JSON numbers decode as float64; enum values are usually ints.
		if f, ok := def.Value.(float64); ok {
			def.Value = int(f)
		}
		if err := l.add(def); err != nil {
			return err
		}
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
		if err := l.add(def); err != nil {
			return err
		}
	}
	return nil
}

// LoadFromSlice loads enum definitions from a slice
func (l *DynamicEnumLoader) LoadFromSlice(definitions []EnumDefinition) error {
	for _, def := range definitions {
		if err := l.add(def); err != nil {
			return err
		}
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
