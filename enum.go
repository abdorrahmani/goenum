package goenum

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Enum represents a basic enum interface
type Enum interface {
	String() string
	Value() interface{}
	IsValid() bool
	Description() string
	HasAlias(alias string) bool
	Aliases() []string
}

// EnumBase provides a basic implementation of Enum interface
type EnumBase struct {
	value       interface{}
	name        string
	description string
	aliases     []string
}

// String returns the string representation of the enum
func (e *EnumBase) String() string {
	if e == nil {
		return ""
	}
	return e.name
}

// Value returns the value of the enum
func (e *EnumBase) Value() interface{} {
	if e == nil {
		return nil
	}
	return e.value
}

// IsValid checks if the enum value is valid
func (e *EnumBase) IsValid() bool {
	return e != nil && e.name != ""
}

// Description returns the description of the enum
func (e *EnumBase) Description() string {
	if e == nil {
		return ""
	}
	return e.description
}

// HasAlias checks if the enum has a specific alias
func (e *EnumBase) HasAlias(alias string) bool {
	if e == nil {
		return false
	}
	for _, a := range e.aliases {
		if strings.EqualFold(a, alias) {
			return true
		}
	}
	return false
}

// Aliases returns all aliases of the enum
func (e *EnumBase) Aliases() []string {
	if e == nil {
		return nil
	}
	return e.aliases
}

// NewEnumSet creates a new EnumSet instance
func NewEnumSet[T Enum]() *EnumSet[T] {
	return &EnumSet[T]{
		values:  make(map[string]T),
		byValue: make(map[interface{}]T),
	}
}

// EnumSet represents a collection of enum values
type EnumSet[T Enum] struct {
	values  map[string]T
	byValue map[interface{}]T
}

// Register adds an enum value to the set and returns the EnumSet for chaining
func (es *EnumSet[T]) Register(enum T) *EnumSet[T] {
	name := enum.String()
	value := enum.Value()

	// Check for duplicate name
	if _, exists := es.values[name]; exists {
		panic(fmt.Sprintf("duplicate enum name: %s", name))
	}

	// Check for duplicate value
	if _, exists := es.byValue[value]; exists {
		panic(fmt.Sprintf("duplicate enum value: %v", value))
	}

	es.values[name] = enum
	es.byValue[value] = enum
	return es
}

// GetByName retrieves an enum by its string name
func (es *EnumSet[T]) GetByName(name string) (T, bool) {
	enum, exists := es.values[strings.ToUpper(name)]
	if exists {
		return enum, true
	}

	// Check aliases
	for _, e := range es.values {
		if e.HasAlias(name) {
			return e, true
		}
	}

	var zero T
	return zero, false
}

// GetByValue retrieves an enum by its value
func (es *EnumSet[T]) GetByValue(value interface{}) (T, bool) {
	enum, exists := es.byValue[value]
	return enum, exists
}

// Values returns all registered enum values
func (es *EnumSet[T]) Values() []T {
	result := make([]T, 0, len(es.values))
	for _, v := range es.values {
		result = append(result, v)
	}
	return result
}

// Contains checks if an enum exists in the set
func (es *EnumSet[T]) Contains(enum T) bool {
	_, exists := es.values[enum.String()]
	return exists
}

// MarshalJSON implements JSON marshaling for enum
func (e *EnumBase) MarshalJSON() ([]byte, error) {
	if e == nil {
		return json.Marshal("")
	}
	return json.Marshal(e.String())
}

// UnmarshalJSON implements JSON unmarshaling for enum
func (e *EnumBase) UnmarshalJSON(data []byte) error {
	if e == nil {
		return fmt.Errorf("cannot unmarshal into nil EnumBase")
	}

	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	e.name = name
	return nil
}

// NewEnumBase creates a new EnumBase with the given parameters
func NewEnumBase(value interface{}, name string, description string, aliases ...string) *EnumBase {
	return &EnumBase{
		value:       value,
		name:        name,
		description: description,
		aliases:     aliases,
	}
}

// Names returns a slice of all enum names in the set
func (es *EnumSet[T]) Names() []string {
	names := make([]string, 0, len(es.values))
	for name := range es.values {
		names = append(names, name)
	}
	return names
}

// Map returns a map of enum names to their values
func (es *EnumSet[T]) Map() map[string]interface{} {
	result := make(map[string]interface{}, len(es.values))
	for name, enum := range es.values {
		result[name] = enum.Value()
	}
	return result
}

// Filter returns a slice of enums that satisfy the given predicate
func (es *EnumSet[T]) Filter(predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, enum := range es.values {
		if predicate(enum) {
			result = append(result, enum)
		}
	}
	return result
}
