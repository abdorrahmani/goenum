package goenum

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Enum represents a basic enum interface.
//
// Legacy: this interface and the EnumBase/EnumSet machinery exist for
// backward compatibility and for dynamic enums (DynamicEnumLoader), where
// values are only known at runtime. For new code prefer the modern API:
// declare a named int/uint/string type, run `goenum generate`, and use the
// generated methods plus goenum.Parse[T] / goenum.Set[T] — no interface{},
// no registration, compile-time type safety.
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
	HasAllFlags(flags ...CompositeEnum) bool
	IsEmpty() bool
	// Flag manipulation
	RemoveFlag(flag CompositeEnum) CompositeEnum
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

// EnumJSONConfig holds configuration for JSON serialization. EnumSet injects
// a fully populated config (including unexported fields) into enums at
// registration time; standalone enums use the exported fields only.
type EnumJSONConfig struct {
	Format JSONFormat

	// resolve resolves a name (applying the owning EnumSet's case and alias
	// rules). nil for standalone enums.
	resolve func(name string) (Enum, bool)
	// unknown is the UnknownBehavior of the owning EnumSet.
	unknown UnknownBehavior
}

// DefaultJSONConfig returns the default JSON configuration
func DefaultJSONConfig() *EnumJSONConfig {
	return &EnumJSONConfig{
		Format: JSONFormatName,
	}
}

// EnumBase provides a basic implementation of Enum interface.
//
// Legacy: generated enums do not embed EnumBase. It remains the runtime
// representation for dynamic enums loaded by DynamicEnumLoader.
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

// NewEnumSet creates a new EnumSet instance. Without options the set uses
// DefaultConfig, preserving v1 behavior: name-format JSON, case-insensitive
// lookups, aliases enabled, unknown names reported as not-found.
func NewEnumSet[T Enum](options ...EnumSetOption) *EnumSet[T] {
	es := &EnumSet[T]{
		values:  make(map[string]T),
		byValue: make(map[interface{}]T),
		config:  DefaultConfig,
	}
	for _, option := range options {
		if option != nil {
			option(&es.config)
		}
	}
	es.config = es.config.normalized()
	es.jsonConfig = &EnumJSONConfig{
		Format: es.config.JSONFormat,
		resolve: func(name string) (Enum, bool) {
			es.mu.RLock()
			defer es.mu.RUnlock()
			return es.lookupName(name)
		},
		unknown: es.config.UnknownBehavior,
	}
	return es
}

// EnumSet represents a collection of enum values.
//
// Legacy: for generated enums use the typed goenum.Set[T], which needs no
// registration and takes typed values. EnumSet remains the container for
// dynamic enums.
type EnumSet[T Enum] struct {
	mu      sync.RWMutex
	values  map[string]T
	byValue map[interface{}]T

	// Lookup maps keyed by normalized name/alias. Written only by Register
	// under the write lock; treat as immutable afterwards for cheap reads.
	byName  map[string]T
	byAlias map[string]T

	config     Config
	jsonConfig *EnumJSONConfig
}

// Config returns a copy of the effective configuration. Configuration is
// immutable after construction; mutating the returned value has no effect on
// the set.
func (es *EnumSet[T]) Config() Config {
	return es.config
}

// normalizeKey maps a name or alias to its lookup key. In case-insensitive
// mode keys are uppercased so "ACTIVE" and "active" share one entry.
func (es *EnumSet[T]) normalizeKey(s string) string {
	if es.config.CaseSensitive {
		return s
	}
	return strings.ToUpper(s)
}

// Register adds an enum value to the set and returns the EnumSet for chaining.
// It panics on duplicate names, duplicate values and ambiguous names or
// aliases (a name colliding with another name or alias under the set's
// case-sensitivity, or an alias shared by two enums).
func (es *EnumSet[T]) Register(enum T) *EnumSet[T] {
	name := enum.String()
	value := enum.Value()

	es.mu.Lock()
	defer es.mu.Unlock()

	// Check for duplicate name
	if _, exists := es.values[name]; exists {
		panic(fmt.Sprintf("duplicate enum name: %s", name))
	}

	// Check for duplicate value
	if _, exists := es.byValue[value]; exists {
		panic(fmt.Sprintf("duplicate enum value: %v", value))
	}

	// Check for case-insensitive collision with an existing name
	nameKey := es.normalizeKey(name)
	if existing, exists := es.byName[nameKey]; exists {
		panic(fmt.Sprintf("ambiguous enum name %q conflicts with name %q (case-insensitive matching is enabled)", name, existing.String()))
	}
	if existing, exists := es.byAlias[nameKey]; exists {
		panic(fmt.Sprintf("ambiguous enum name %q conflicts with alias %q of enum %q", name, nameKey, existing.String()))
	}

	es.values[name] = enum
	es.byValue[value] = enum
	if es.byName == nil {
		es.byName = make(map[string]T)
	}
	es.byName[nameKey] = enum

	// Index aliases. Aliases stay part of the enum metadata even when
	// AllowAliases is disabled; only resolution is gated.
	if es.config.AllowAliases {
		for _, alias := range enum.Aliases() {
			aliasKey := es.normalizeKey(alias)
			if aliasKey == "" {
				continue
			}
			if existing, exists := es.byName[aliasKey]; exists && existing.String() != name {
				panic(fmt.Sprintf("ambiguous alias %q of enum %q conflicts with name of enum %q", alias, name, existing.String()))
			}
			if existing, exists := es.byAlias[aliasKey]; exists && existing.String() != name {
				panic(fmt.Sprintf("ambiguous alias %q: used by both enum %q and enum %q", alias, existing.String(), name))
			}
			if es.byAlias == nil {
				es.byAlias = make(map[string]T)
			}
			es.byAlias[aliasKey] = enum
		}
	}

	if es.jsonConfig != nil {
		injectJSONConfig(enum, es.jsonConfig)
	}
	return es
}

// injectJSONConfig sets the shared EnumJSONConfig on enums that embed
// *EnumBase. Enums with their own SetJSONConfig override keep their choice.
func injectJSONConfig(enum Enum, config *EnumJSONConfig) {
	if base, ok := enum.(interface{ SetJSONConfig(*EnumJSONConfig) }); ok {
		base.SetJSONConfig(config)
	}
}

// lookupName resolves a name or alias according to the set's case and alias
// configuration. The caller chooses whether failure is an error.
func (es *EnumSet[T]) lookupName(name string) (T, bool) {
	if es.config.CaseSensitive {
		if enum, exists := es.byName[name]; exists {
			return enum, true
		}
	} else if enum, exists := es.byName[strings.ToUpper(name)]; exists {
		return enum, true
	}
	if !es.config.AllowAliases {
		var zero T
		return zero, false
	}
	if es.config.CaseSensitive {
		enum, exists := es.byAlias[name]
		return enum, exists
	}
	enum, exists := es.byAlias[strings.ToUpper(name)]
	return enum, exists
}

// GetByName retrieves an enum by its string name. Name matching follows the
// set's CaseSensitive configuration; alias resolution follows AllowAliases.
func (es *EnumSet[T]) GetByName(name string) (T, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.lookupName(name)
}

// GetByValue retrieves an enum by its value
func (es *EnumSet[T]) GetByValue(value interface{}) (T, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()
	enum, exists := es.byValue[value]
	return enum, exists
}

// Parse resolves a name (or alias when allowed) to a registered enum,
// following the set's UnknownBehavior:
//
//   - UnknownError: returns an error for unknown names.
//   - UnknownZero:  returns the zero enum without error.
//   - UnknownIgnore: behaves like UnknownZero.
func (es *EnumSet[T]) Parse(name string) (T, error) {
	var zero T
	enum, exists := es.lookupName(name)
	if exists {
		return enum, nil
	}
	switch es.config.UnknownBehavior {
	case UnknownZero, UnknownIgnore:
		return zero, nil
	default: // UnknownError
		return zero, fmt.Errorf("unknown enum name: %s", name)
	}
}

// Values returns all registered enum values
func (es *EnumSet[T]) Values() []T {
	es.mu.RLock()
	defer es.mu.RUnlock()
	result := make([]T, 0, len(es.values))
	for _, v := range es.values {
		result = append(result, v)
	}
	return result
}

// Contains checks if an enum exists in the set
func (es *EnumSet[T]) Contains(enum T) bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
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
		// Enums registered in a configured EnumSet resolve and validate
		// names against the set; standalone enums keep v1 behavior.
		if config.resolve != nil {
			resolved, ok := config.resolve(name)
			if !ok {
				switch config.unknown {
				case UnknownZero:
					e.name, e.value, e.description, e.aliases = "", nil, "", nil
					return nil
				case UnknownIgnore:
					return nil
				default: // UnknownError
					return fmt.Errorf("unknown enum name: %s", name)
				}
			}
			e.name = resolved.String()
			e.value = resolved.Value()
			e.description = resolved.Description()
			e.aliases = resolved.Aliases()
			return nil
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
	es.mu.RLock()
	defer es.mu.RUnlock()
	names := make([]string, 0, len(es.values))
	for name := range es.values {
		names = append(names, name)
	}
	return names
}

// Map returns a map of enum names to their values
func (es *EnumSet[T]) Map() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()
	result := make(map[string]interface{}, len(es.values))
	for name, enum := range es.values {
		result[name] = enum.Value()
	}
	return result
}

// Filter returns a slice of enums that satisfy the given predicate. The
// predicate runs outside the set's lock and must not register enums.
func (es *EnumSet[T]) Filter(predicate func(T) bool) []T {
	es.mu.RLock()
	all := make([]T, 0, len(es.values))
	for _, enum := range es.values {
		all = append(all, enum)
	}
	es.mu.RUnlock()
	result := make([]T, 0)
	for _, enum := range all {
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

// HasAllFlags checks if all given flags are present in the composite enum
func (e *CompositeEnumBase) HasAllFlags(flags ...CompositeEnum) bool {
	if e == nil || len(flags) == 0 {
		return false
	}
	for _, flag := range flags {
		if !e.HasFlag(flag) {
			return false
		}
	}
	return true
}

// RemoveFlag removes a specific flag from the composite enum
func (e *CompositeEnumBase) RemoveFlag(flag CompositeEnum) CompositeEnum {
	if e == nil || flag == nil {
		return e
	}
	flagBase, ok := flag.(*CompositeEnumBase)
	if !ok {
		return e
	}
	newFlags := e.flags &^ flagBase.flags
	return &CompositeEnumBase{
		EnumBase: NewEnumBase(newFlags, e.name+"-"+flag.String(), e.description),
		flags:    newFlags,
	}
}
