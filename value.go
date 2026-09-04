package goenum

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Value constrains the underlying types supported by the modern, generated
// enum API. Any named type whose underlying type is one of these basic types
// can have enum boilerplate generated for it:
//
//	type Status int
//	type Priority string
//	type Permission uint64
type Value interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~string
}

// ErrUnknownEnum is returned (wrapped) by Parse helpers and generated
// Unmarshal methods when a name does not resolve to a known enum value.
var ErrUnknownEnum = errors.New("goenum: unknown enum name")

// EnumInfo is the immutable metadata table for one generated enum type. The
// goenum generator emits a package-level variable built from a literal
// EnumInfo and registered via Register; generated methods then perform O(1)
// map lookups against it. EnumInfo is safe for concurrent use after
// registration and must not be mutated.
type EnumInfo[T Value] struct {
	// Names holds the canonical name of each value, index-aligned with Values.
	Names []string
	// Values holds every enum value in declaration order.
	Values []T
	// Descs holds the description of each value, index-aligned with Values.
	// It may be nil when no descriptions are declared.
	Descs []string
	// Aliases holds the aliases of each value, index-aligned with Values.
	// Individual entries may be nil.
	Aliases [][]string

	// Lookup maps. Built once by Register.
	exactName  map[string]T
	foldName   map[string]T
	exactAlias map[string]T
	foldAlias  map[string]T
	nameByVal  map[T]string
	descByVal  map[T]string
	aliasByVal map[T][]string
}

// Register builds the lookup tables of info, records it in the global type
// registry so the generic helpers (Parse, MustParse, Name, Valid, Values,
// Names) and Set[T] can find it, and returns info. Generated code calls it
// exactly once per enum type, usually in a package-level variable
// initializer. Register panics if info is malformed (length mismatch,
// duplicate names, values or aliases) or if the type is already registered.
func Register[T Value](info EnumInfo[T]) *EnumInfo[T] {
	if len(info.Names) != len(info.Values) {
		panic(fmt.Sprintf("goenum: EnumInfo[%T]: %d names but %d values", *new(T), len(info.Names), len(info.Values)))
	}
	if info.Descs != nil && len(info.Descs) != len(info.Values) {
		panic(fmt.Sprintf("goenum: EnumInfo[%T]: %d descriptions but %d values", *new(T), len(info.Descs), len(info.Values)))
	}
	if info.Aliases != nil && len(info.Aliases) != len(info.Values) {
		panic(fmt.Sprintf("goenum: EnumInfo[%T]: %d alias lists but %d values", *new(T), len(info.Aliases), len(info.Values)))
	}

	built := EnumInfo[T]{
		Names:      info.Names,
		Values:     info.Values,
		Descs:      info.Descs,
		Aliases:    info.Aliases,
		exactName:  make(map[string]T, len(info.Names)),
		foldName:   make(map[string]T, len(info.Names)),
		nameByVal:  make(map[T]string, len(info.Values)),
		descByVal:  make(map[T]string, len(info.Values)),
		aliasByVal: make(map[T][]string, len(info.Values)),
	}

	for i, v := range info.Values {
		name := info.Names[i]
		if _, dup := built.nameByVal[v]; dup {
			panic(fmt.Sprintf("goenum: EnumInfo[%T]: duplicate value %v (name %q)", v, v, name))
		}
		if _, dup := built.exactName[name]; dup {
			panic(fmt.Sprintf("goenum: EnumInfo[%T]: duplicate name %q", v, name))
		}
		fold := strings.ToUpper(name)
		if _, dup := built.foldName[fold]; dup {
			panic(fmt.Sprintf("goenum: EnumInfo[%T]: ambiguous name %q (case-insensitive collision)", v, name))
		}
		built.exactName[name] = v
		built.foldName[fold] = v
		built.nameByVal[v] = name
		if info.Descs != nil && info.Descs[i] != "" {
			built.descByVal[v] = info.Descs[i]
		}
	}

	for i, v := range info.Values {
		if info.Aliases == nil {
			break
		}
		for _, alias := range info.Aliases[i] {
			if alias == "" {
				continue
			}
			if _, dup := built.exactName[alias]; dup {
				panic(fmt.Sprintf("goenum: EnumInfo[%T]: alias %q of value %v collides with a name", v, alias, v))
			}
			if _, dup := built.exactAlias[alias]; dup {
				panic(fmt.Sprintf("goenum: EnumInfo[%T]: duplicate alias %q", v, alias))
			}
			fold := strings.ToUpper(alias)
			if _, dup := built.foldName[fold]; dup && fold != strings.ToUpper(built.nameByVal[v]) {
				panic(fmt.Sprintf("goenum: EnumInfo[%T]: alias %q collides case-insensitively with a name", v, alias))
			}
			if built.exactAlias == nil {
				built.exactAlias = make(map[string]T)
				built.foldAlias = make(map[string]T)
			}
			if prev, dup := built.foldAlias[fold]; dup && prev != v {
				panic(fmt.Sprintf("goenum: EnumInfo[%T]: alias %q is ambiguous (case-insensitive collision across values)", v, alias))
			}
			built.exactAlias[alias] = v
			built.foldAlias[fold] = v
			built.aliasByVal[v] = append(built.aliasByVal[v], alias)
		}
	}

	registerInfo(reflect.TypeOf((*T)(nil)).Elem(), &built)
	return &built
}

// Lookup resolves a canonical name or alias to a value. Matching is
// case-insensitive and aliases are accepted, matching the library defaults.
func (i *EnumInfo[T]) Lookup(name string) (T, bool) {
	return i.LookupOpts(name, false, true)
}

// LookupOpts resolves a name (and, when allowAliases is set, an alias) with
// the given case-sensitivity. It is used by Set[T] to honour Config.
func (i *EnumInfo[T]) LookupOpts(name string, caseSensitive, allowAliases bool) (T, bool) {
	var zero T
	if caseSensitive {
		if v, ok := i.exactName[name]; ok {
			return v, true
		}
		if allowAliases && i.exactAlias != nil {
			v, ok := i.exactAlias[name]
			return v, ok
		}
		return zero, false
	}
	fold := strings.ToUpper(name)
	if v, ok := i.foldName[fold]; ok {
		return v, true
	}
	if allowAliases && i.foldAlias != nil {
		v, ok := i.foldAlias[fold]
		return v, ok
	}
	return zero, false
}

// Name returns the canonical name of v and whether v is a known value.
func (i *EnumInfo[T]) Name(v T) (string, bool) {
	name, ok := i.nameByVal[v]
	return name, ok
}

// Description returns the description of v, or "" when v is unknown or has no
// description.
func (i *EnumInfo[T]) Description(v T) string { return i.descByVal[v] }

// AliasesOf returns the aliases declared for v (never mutated by callers).
func (i *EnumInfo[T]) AliasesOf(v T) []string { return i.aliasByVal[v] }

// HasAlias reports whether alias (compared case-insensitively) belongs to v.
func (i *EnumInfo[T]) HasAlias(v T, alias string) bool {
	for _, a := range i.aliasByVal[v] {
		if strings.EqualFold(a, alias) {
			return true
		}
	}
	return false
}

// All returns the values in declaration order. The returned slice is shared;
// do not mutate it.
func (i *EnumInfo[T]) All() []T { return i.Values }

// AllNames returns the canonical names in declaration order. The returned
// slice is shared; do not mutate it.
func (i *EnumInfo[T]) AllNames() []string { return i.Names }

var (
	registryMu sync.RWMutex
	registry   = map[reflect.Type]any{}
)

func registerInfo(t reflect.Type, info any) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if prev, exists := registry[t]; exists {
		panic(fmt.Sprintf("goenum: enum type %s is already registered (%v)", t, reflect.ValueOf(prev).Type()))
	}
	registry[t] = info
}

// Info returns the registered EnumInfo for the enum type T, and whether T was
// registered (generated code registers its type automatically).
func Info[T Value]() (*EnumInfo[T], bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	info, ok := registry[reflect.TypeOf((*T)(nil)).Elem()].(*EnumInfo[T])
	return info, ok
}

func mustInfo[T Value]() *EnumInfo[T] {
	info, ok := Info[T]()
	if !ok {
		panic(fmt.Sprintf("goenum: type %v is not a registered enum; run `goenum generate` for it", reflect.TypeOf((*T)(nil)).Elem()))
	}
	return info
}

// Parse resolves name to a value of the generated enum type T, case-insensitively
// and including aliases. It returns an error wrapping ErrUnknownEnum when the
// name is unknown, and panics when T has no generated registration.
func Parse[T Value](name string) (T, error) {
	v, ok := mustInfo[T]().Lookup(name)
	if !ok {
		var zero T
		return zero, fmt.Errorf("%w: %q for %v", ErrUnknownEnum, name, reflect.TypeOf(zero))
	}
	return v, nil
}

// MustParse is Parse but panics on unknown names. Use it only for inputs known
// to be valid, such as constants and test fixtures.
func MustParse[T Value](name string) T {
	v, err := Parse[T](name)
	if err != nil {
		panic(err)
	}
	return v
}

// Name returns the canonical name of v, or "" when v is not a known value of
// the generated enum type T.
func Name[T Value](v T) string {
	name, _ := mustInfo[T]().Name(v)
	return name
}

// Valid reports whether v is one of the declared values of the generated enum
// type T.
func Valid[T Value](v T) bool {
	_, ok := mustInfo[T]().Name(v)
	return ok
}

// Values returns every declared value of the generated enum type T in
// declaration order. The returned slice is shared; do not mutate it.
func Values[T Value]() []T { return mustInfo[T]().All() }

// Names returns every canonical name of the generated enum type T in
// declaration order. The returned slice is shared; do not mutate it.
func Names[T Value]() []string { return mustInfo[T]().AllNames() }
