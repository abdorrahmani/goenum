package goenum

import (
	"encoding/json"
	"strings"
)

// Enum represents a basic enum interface
type Enum interface {
	String() string
	Value() int
	IsValid() bool
}

// EnumBase provides a basic implementation of Enum interface
type EnumBase struct {
	value int
	name  string
}

// String returns the string representation of the enum
func (e *EnumBase) String() string {
	if e == nil {
		return ""
	}
	return e.name
}

// Value returns the numeric value of the enum
func (e *EnumBase) Value() int {
	if e == nil {
		return 0
	}
	return e.value
}

// IsValid checks if the enum value is valid
func (e *EnumBase) IsValid() bool {
	return e != nil && e.name != ""
}

// NewEnumSet creates a new EnumSet instance
func NewEnumSet[T Enum]() *EnumSet[T] {
	return &EnumSet[T]{
		values: make(map[string]T),
	}
}

// EnumSet represents a collection of enum values
type EnumSet[T Enum] struct {
	values map[string]T
}

// Register adds an enum value to the set
func (es *EnumSet[T]) Register(enum T) {
	es.values[enum.String()] = enum
}

// GetByName retrieves an enum by its string name
func (es *EnumSet[T]) GetByName(name string) (T, bool) {
	enum, exists := es.values[strings.ToUpper(name)]
	return enum, exists
}

// GetByValue retrieves an enum by its integer value
func (es *EnumSet[T]) GetByValue(value int) (T, bool) {
	for _, enum := range es.values {
		if enum.Value() == value {
			return enum, true
		}
	}
	var zero T
	return zero, false
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
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	*e = EnumBase{name: name}
	return nil
}
