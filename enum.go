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

// CompositeEnum represents an enum that can be combined with other enums
type CompositeEnum interface {
	Enum
	// Bitwise operations
	Or(other CompositeEnum) CompositeEnum
	And(other CompositeEnum) CompositeEnum
	Xor(other CompositeEnum) CompositeEnum
	Not() CompositeEnum
	// Checks
	HasFlag(flag CompositeEnum) bool
	IsEmpty() bool
}

// JSONFormat defines how an enum should be serialized to JSON
type JSONFormat int

const (
	// JSONFormatName serializes only the enum name (default)
	JSONFormatName JSONFormat = iota
	// JSONFormatValue serializes only the enum value
	JSONFormatValue
	// JSONFormatFull serializes a complete struct with all enum information
	JSONFormatFull
)

// EnumJSONConfig holds configuration for JSON serialization
type EnumJSONConfig struct {
	Format JSONFormat
}

// DefaultJSONConfig returns the default JSON configuration
func DefaultJSONConfig() *EnumJSONConfig {
	return &EnumJSONConfig{
		Format: JSONFormatName,
	}
}

// EnumBase provides a basic implementation of Enum interface
type EnumBase struct {
	value       interface{}
	name        string
	description string
	aliases     []string
	jsonConfig  *EnumJSONConfig
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

// SetJSONConfig sets the JSON serialization configuration
func (e *EnumBase) SetJSONConfig(config *EnumJSONConfig) {
	if e == nil {
		return
	}
	e.jsonConfig = config
}

// GetJSONConfig returns the current JSON configuration
func (e *EnumBase) GetJSONConfig() *EnumJSONConfig {
	if e == nil || e.jsonConfig == nil {
		return DefaultJSONConfig()
	}
	return e.jsonConfig
}

// MarshalJSON implements JSON marshaling for enum
func (e *EnumBase) MarshalJSON() ([]byte, error) {
	if e == nil {
		return json.Marshal("")
	}

	config := e.GetJSONConfig()
	switch config.Format {
	case JSONFormatValue:
		return json.Marshal(e.Value())
	case JSONFormatFull:
		type FullEnum struct {
			Name        string      `json:"name"`
			Value       interface{} `json:"value"`
			Description string      `json:"description"`
			Aliases     []string    `json:"aliases,omitempty"`
		}
		return json.Marshal(FullEnum{
			Name:        e.name,
			Value:       e.value,
			Description: e.description,
			Aliases:     e.aliases,
		})
	default: // JSONFormatName
		return json.Marshal(e.String())
	}
}

// UnmarshalJSON implements JSON unmarshaling for enum
func (e *EnumBase) UnmarshalJSON(data []byte) error {
	if e == nil {
		return fmt.Errorf("cannot unmarshal into nil EnumBase")
	}

	config := e.GetJSONConfig()
	switch config.Format {
	case JSONFormatValue:
		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		// Convert float64 to int if necessary
		if f, ok := value.(float64); ok {
			e.value = int(f)
		} else {
			e.value = value
		}
		return nil
	case JSONFormatFull:
		type FullEnum struct {
			Name        string      `json:"name"`
			Value       interface{} `json:"value"`
			Description string      `json:"description"`
			Aliases     []string    `json:"aliases,omitempty"`
		}
		var full FullEnum
		if err := json.Unmarshal(data, &full); err != nil {
			return err
		}
		e.name = full.Name
		// Convert float64 to int if necessary
		if f, ok := full.Value.(float64); ok {
			e.value = int(f)
		} else {
			e.value = full.Value
		}
		e.description = full.Description
		e.aliases = full.Aliases
		return nil
	default: // JSONFormatName
		var name string
		if err := json.Unmarshal(data, &name); err != nil {
			return err
		}
		e.name = name
		return nil
	}
}

// NewEnumBase creates a new EnumBase with the given parameters
func NewEnumBase(value interface{}, name string, description string, aliases ...string) *EnumBase {
	return &EnumBase{
		value:       value,
		name:        name,
		description: description,
		aliases:     aliases,
		jsonConfig:  DefaultJSONConfig(),
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

// CompositeEnumBase provides a basic implementation of CompositeEnum interface
type CompositeEnumBase struct {
	*EnumBase
	flags uint64
}

// NewCompositeEnumBase creates a new CompositeEnumBase with the given parameters
func NewCompositeEnumBase(value interface{}, name string, description string, aliases ...string) *CompositeEnumBase {
	flags, ok := value.(uint64)
	if !ok {
		// If value is not uint64, use 1 << value as the flag
		if intVal, ok := value.(int); ok {
			flags = 1 << uint(intVal)
		} else {
			flags = 0
		}
	}
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(flags, name, description, aliases...),
		flags:    flags,
	}
}

// Or performs a bitwise OR operation with another enum
func (e *CompositeEnumBase) Or(other CompositeEnum) CompositeEnum {
	if e == nil || other == nil {
		return e
	}
	otherBase, ok := other.(*CompositeEnumBase)
	if !ok {
		return e
	}
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(e.flags|otherBase.flags, e.name+"|"+other.String(), e.description),
		flags:    e.flags | otherBase.flags,
	}
}

// And performs a bitwise AND operation with another enum
func (e *CompositeEnumBase) And(other CompositeEnum) CompositeEnum {
	if e == nil || other == nil {
		return e
	}
	otherBase, ok := other.(*CompositeEnumBase)
	if !ok {
		return e
	}
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(e.flags&otherBase.flags, e.name+"&"+other.String(), e.description),
		flags:    e.flags & otherBase.flags,
	}
}

// Xor performs a bitwise XOR operation with another enum
func (e *CompositeEnumBase) Xor(other CompositeEnum) CompositeEnum {
	if e == nil || other == nil {
		return e
	}
	otherBase, ok := other.(*CompositeEnumBase)
	if !ok {
		return e
	}
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(e.flags^otherBase.flags, e.name+"^"+other.String(), e.description),
		flags:    e.flags ^ otherBase.flags,
	}
}

// Not performs a bitwise NOT operation
func (e *CompositeEnumBase) Not() CompositeEnum {
	if e == nil {
		return e
	}
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(^e.flags, "~"+e.name, e.description),
		flags:    ^e.flags,
	}
}

// HasFlag checks if the enum has a specific flag set
func (e *CompositeEnumBase) HasFlag(flag CompositeEnum) bool {
	if e == nil || flag == nil {
		return false
	}
	flagBase, ok := flag.(*CompositeEnumBase)
	if !ok {
		return false
	}
	return (e.flags & flagBase.flags) == flagBase.flags
}

// IsEmpty checks if the enum has no flags set
func (e *CompositeEnumBase) IsEmpty() bool {
	return e == nil || e.flags == 0
}

// Value returns the flags value
func (e *CompositeEnumBase) Value() interface{} {
	if e == nil {
		return nil
	}
	return e.flags
}
