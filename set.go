package goenum

import "fmt"

// Set is the typed, configuration-aware view over a generated enum type. It
// replaces the legacy EnumSet[T Enum] for generated enums: lookups are
// value-typed (no interface{}), and the same Config options that govern
// EnumSet (case sensitivity, aliases, unknown-name behavior) apply here.
//
// The enum data itself comes from the type's registered EnumInfo, so a Set is
// just a policy wrapper — creating one is cheap and there is no registration
// step:
//
//	Statuses := goenum.NewSet[Status]()
//	s, err := Statuses.Parse("RUNNING") // alias of ACTIVE
//
// Set is safe for concurrent use; it holds no mutable state beyond its
// immutable Config and the shared EnumInfo.
type Set[T Value] struct {
	info   *EnumInfo[T]
	config Config
}

// NewSet creates a Set for the generated enum type T. Without options it uses
// DefaultConfig, matching the legacy EnumSet defaults (case-insensitive,
// aliases enabled, unknown names reported as errors). It panics if T has no
// generated registration.
func NewSet[T Value](options ...EnumSetOption) *Set[T] {
	// Options are applied on top of DefaultConfig so partial configuration
	// (e.g. only WithCaseSensitive) keeps the v1-compatible defaults.
	cfg := DefaultConfig
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}
	return &Set[T]{
		info:   mustInfo[T](),
		config: cfg.normalized(),
	}
}

// Config returns a copy of the effective configuration.
func (s *Set[T]) Config() Config { return s.config }

// GetByName retrieves a value by canonical name; alias resolution and case
// sensitivity follow the set's Config.
func (s *Set[T]) GetByName(name string) (T, bool) {
	return s.info.LookupOpts(name, s.config.CaseSensitive, s.config.AllowAliases)
}

// GetByValue retrieves the canonical name of a value. Unlike the legacy
// EnumSet.GetByValue, the parameter is typed: wrong-type lookups are compile
// errors.
func (s *Set[T]) GetByValue(value T) (string, bool) {
	return s.info.Name(value)
}

// Parse resolves a name (or alias when enabled) following the set's
// UnknownBehavior, exactly like (*EnumSet[T]).Parse.
func (s *Set[T]) Parse(name string) (T, error) {
	v, ok := s.GetByName(name)
	if ok {
		return v, nil
	}
	var zero T
	switch s.config.UnknownBehavior {
	case UnknownZero, UnknownIgnore:
		return zero, nil
	default: // UnknownError
		return zero, fmt.Errorf("unknown enum name: %s", name)
	}
}

// MustParse is Parse but panics on unknown names.
func (s *Set[T]) MustParse(name string) T {
	v, err := s.Parse(name)
	if err != nil {
		panic(err)
	}
	return v
}

// Contains reports whether value is one of the declared enum values.
func (s *Set[T]) Contains(value T) bool {
	_, ok := s.info.Name(value)
	return ok
}

// Values returns all declared values in declaration order. The returned slice
// is shared; do not mutate it.
func (s *Set[T]) Values() []T { return s.info.All() }

// Names returns all canonical names in declaration order. The returned slice
// is shared; do not mutate it.
func (s *Set[T]) Names() []string { return s.info.AllNames() }

// Description returns the description of value, or "" when unknown or
// undescribed.
func (s *Set[T]) Description(value T) string { return s.info.Description(value) }

// Aliases returns the aliases declared for value.
func (s *Set[T]) Aliases(value T) []string { return s.info.AliasesOf(value) }
