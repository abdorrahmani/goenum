# GoEnum - Type-Safe Enums for Go

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](https://golang.org/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/abdorrahmani/goenum)](https://goreportcard.com/report/github.com/abdorrahmani/goenum)
[![License: MIT](https://img.shields.io/github/license/abdorrahmani/goenum?logo=open-source-initiative&logoColor=white)](https://github.com/abdorrahmani/goenum/blob/main/LICENSE)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/abdorrahmani/goenum)

GoEnum is a production-grade enum toolkit for Go. Declare enums as plain
named `int`/`uint`/`string` types, run the generator, and get strongly typed
`String`, `Parse`, JSON, text and SQL support with **zero reflection and zero
allocations on hot paths**.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [The Generator](#the-generator)
  - [Declaration Directives](#declaration-directives)
  - [CLI](#cli)
  - [Generated API](#generated-api)
- [Generic Helpers](#generic-helpers)
- [Typed Sets](#typed-sets)
- [JSON, Text and SQL](#json-text-and-sql)
- [Bit Flags](#bit-flags)
- [Configuration](#configuration)
- [Dynamic Enums](#dynamic-enums)
- [Legacy API (EnumBase / EnumSet)](#legacy-api-enumbase--enumset)
- [Migrating from the Legacy API](#migrating-from-the-legacy-api)
- [Performance](#performance)
- [Contributing](#contributing)
- [License](#license)

## 🚀 Installation

```bash
go get github.com/abdorrahmani/goenum
```

Install the generator CLI:

```bash
go install github.com/abdorrahmani/goenum/cmd/goenum@latest
```

**Requirements:** Go 1.21 or higher (see `go.mod`).

## 🎯 Quick Start

```go
package domain

//go:generate goenum generate

// Status represents the lifecycle state of an application.
type Status int

const (
	//goenum:name=PENDING
	//goenum:description=Waiting to be processed
	//goenum:alias=WAITING
	StatusPending Status = iota
	//goenum:name=ACTIVE
	//goenum:description=Currently active
	//goenum:alias=RUNNING
	StatusActive
	//goenum:name=DELETED
	//goenum:description=The item has been deleted
	//goenum:alias=REMOVED
	StatusDeleted
)
```

Run `go generate ./...` (or `goenum generate`). You get, with no hand-written
boilerplate:

```go
StatusActive.String()            // "ACTIVE"
StatusActive.IsValid()           // true
StatusActive.Description()       // "Currently active"
StatusActive.Aliases()           // ["RUNNING"]

status.ParseStatus("RUNNING")    // StatusActive, nil  (aliases, case-insensitive)
status.MustParseStatus("ACTIVE") // StatusActive       (panics on unknown names)
status.StatusFromValue(1)        // StatusActive, true

goenum.Parse[Status]("active")   // StatusActive, nil  (generic helper)
goenum.Name(StatusDeleted)       // "DELETED"

json.Marshal(StatusActive)       // "ACTIVE"
```

No `EnumBase`, no registration, no `interface{}` — `Status` stays a plain
`int` you can compare, store in slices, and use as a map key.

## The Generator

The generator reads your source with `go/ast`, evaluates the declared
constants (including `iota` and `1 << iota`), validates them, and writes one
`<type>_enum_gen.go` file per enum next to the source. Output is gofmt'd,
deterministic, and contains static lookup tables — no reflection at runtime.

### Declaration Directives

Metadata is declared with `//goenum:` comments, so the syntax is
formatter-friendly and needs no custom DSL:

| Directive | Where | Meaning |
|---|---|---|
| `//goenum:enum` | type | Opt the type into generation without per-const directives (names default to the identifier with the type prefix stripped and uppercased) |
| `//goenum:name=X` | const | Canonical name. Defaults to the identifier with the type prefix stripped and uppercased (`StatusPending` → `PENDING`) |
| `//goenum:description=T` | const | Human-readable description |
| `//goenum:alias=X` | const | Additional parseable name; repeat the directive for several aliases |
| `//goenum:ignore` | const | Skip this constant (e.g. sentinels) |
| `//goenum:flags` | type | Treat the type as a bit-flag set (unsigned underlying type only) |

Generation is **opt-in**: a type is only treated as an enum when it carries
`//goenum:enum` or `//goenum:flags`, or when at least one of its constants
carries a `//goenum:` directive. Plain `int` types with constants are left
alone, so `goenum generate ./...` is safe to run over a whole repository.

Supported underlying types: `int`, `int8`–`int64`, `uint`, `uint8`–`uint64`,
`string`. Values may be literals, `iota` expressions, `1 << iota`, or simple
arithmetic over earlier constants (including plain package constants).

### CLI

```bash
goenum generate [path...]        # files, directories, or dir/... (default: .)
goenum generate ./...            # whole repository (only opted-in enums)
goenum generate --dry-run ./...  # report without writing
goenum generate --force <path>   # overwrite a non-generated file at the output path
```

The generator validates declarations and fails with positioned errors
(`file:line: Type: reason`) for duplicate names, duplicate values, duplicate
aliases, alias/name conflicts, non-power-of-two flags, unsupported types and
invalid directives.

### Generated API

For each enum `T` the generator emits:

- Methods: `String`, `IsValid`, `Description`, `Aliases`, `HasAlias`
- Functions: `ParseT(string) (T, error)`, `MustParseT(string) T`,
  `TFromValue(v) (T, bool)`
- Interfaces: `MarshalJSON`/`UnmarshalJSON` (name format),
  `MarshalText`/`UnmarshalText`, `driver.Valuer`/`sql.Scanner`

`Must...` functions **panic** on invalid input — use them only for constants
and fixtures.

## Generic Helpers

The generated code registers its lookup table automatically, so the generic
helpers work on any generated enum:

```go
status, err := goenum.Parse[Status]("ACTIVE")
status := goenum.MustParse[Status]("ACTIVE")
goenum.Valid[Status](status)
goenum.Name(status)
goenum.Values[Status]()   // all values, declaration order
goenum.Names[Status]()    // all canonical names
```

These are convenience wrappers over the same tables the generated methods
use; prefer the type-specific `ParseStatus` when you want package-local code
with no import of the runtime helpers.

## Typed Sets

`goenum.Set[T]` is a configuration-aware view over a generated enum — typed
lookups, no registration:

```go
Statuses := goenum.NewSet[Status]()

Statuses.GetByName("ACTIVE")        // (Status, bool)
Statuses.GetByValue(StatusActive)   // (string, bool) — typed, not interface{}
Statuses.Parse("RUNNING")           // (Status, error)
Statuses.Contains(StatusDeleted)    // bool
Statuses.Values()                   // []Status
Statuses.Names()                    // []string
Statuses.Description(StatusActive)  // string
```

## JSON, Text and SQL

```go
// JSON — name format, aliases accepted on input
data, _ := json.Marshal(StatusActive)   // "ACTIVE"
var s Status
json.Unmarshal([]byte(`"RUNNING"`), &s) // StatusActive

// Text — works with query params, forms, YAML, logging
text, _ := StatusActive.MarshalText()   // "ACTIVE"
s.UnmarshalText([]byte("active"))       // case-insensitive

// SQL — stores the underlying value, scans names or values
db.Exec("INSERT INTO users (status) VALUES (?)", StatusActive)
db.QueryRow(...).Scan(&s)
```

Undeclared values are a marshaling error rather than silent garbage.

## Bit Flags

```go
//goenum:flags
type Permission uint64

const (
	PermissionRead Permission = 1 << iota
	PermissionWrite
	PermissionDelete
)
```

Generated helpers:

```go
p := PermissionRead.Add(PermissionWrite)
p.Has(PermissionRead)          // true
p.Remove(PermissionWrite)      // PermissionRead
p.String()                     // "READ|WRITE"
ParsePermission("READ|DELETE") // combined value
p.IsValid()                    // false if any undeclared bit is set
```

Flags are deliberately separate from scalar enums: a `//goenum:flags` type
gets `Has`/`Add`/`Remove`/`IsEmpty` and pipe-joined parsing, not scalar
semantics.

## Configuration

Behavior of `Set[T]` (and the legacy `EnumSet`) is controlled by a single
`goenum.Config` applied at construction time. Configuration is immutable
afterwards; there is no global mutable state.

```go
Statuses := goenum.NewSet[Status](
	goenum.WithCaseSensitive(false),
	goenum.WithAliases(true),
	goenum.WithUnknownBehavior(goenum.UnknownError),
)
```

| Field | Meaning |
|---|---|
| `JSONFormat` | Legacy `EnumSet` JSON output: `JSONFormatName`, `JSONFormatValue`, `JSONFormatFull`. Generated enums always marshal as names |
| `CaseSensitive` | `true`: `ACTIVE != active`. `false`: matching ignores case |
| `AllowAliases` | Whether aliases resolve during lookups and `Parse` |
| `UnknownBehavior` | `UnknownError` (default), `UnknownZero` (zero value, no error), `UnknownIgnore` (like `UnknownZero`; JSON unmarshal leaves target unchanged) |

`goenum.DefaultConfig` preserves v1 behavior: name-format JSON,
case-insensitive lookups, aliases enabled, unknown names reported as errors.

## Dynamic Enums

Enums whose values are only known at runtime (loaded from JSON, maps or
slices) keep using the legacy machinery — they cannot be code-generated, and
that is intentional:

```go
loader := goenum.NewDynamicEnumLoader(nil)
err := loader.LoadFromJSON("enums.json")
set := loader.GetEnumSet() // *goenum.EnumSet[goenum.Enum]
```

`LoadFromJSON`, `LoadFromReader`, `LoadFromDirectory`, `LoadFromMap`,
`LoadFromSlice` and `ExportToJSON` are supported, with `ValidationOptions`
for duplicate handling (`DuplicateError`, `DuplicateSkip`,
`DuplicateOverride`), value-type restrictions and empty-name/value policy.

Example JSON format:

```json
[
  { "name": "TEST_A", "value": 1, "description": "Test enum A", "aliases": ["ALPHA"] }
]
```

## Legacy API (EnumBase / EnumSet)

The original architecture remains fully supported — it is still the right
tool for dynamic enums and for existing code:

```go
type Status struct {
	*goenum.EnumBase
}

var (
	StatusPending = Status{goenum.NewEnumBase(0, "PENDING", "Waiting to be processed", "WAITING")}
	StatusActive  = Status{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING")}
)

var Statuses = goenum.NewEnumSet[Status]().
	Register(StatusPending).
	Register(StatusActive)

Statuses.GetByName("ACTIVE") // (Status, bool)
Statuses.GetByValue(1)       // (Status, bool)
Statuses.Parse("RUNNING")    // (Status, error)
```

Composite enums (`NewCompositeEnumBase`, `Or`/`And`/`Xor`/`Not`/`HasFlag`/
`IsEmpty`) and the reflection helpers (`GetEnumMetadata`, `EnumReflection`,
…) are likewise unchanged.

## Migrating from the Legacy API

Before:

```go
type Status struct{ *goenum.EnumBase }

var StatusActive = Status{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING")}
var Statuses = goenum.NewEnumSet[Status]().Register(StatusActive)

s, _ := Statuses.GetByName("RUNNING")
v := s.Value() // interface{}
```

After:

```go
type Status int

const (
	//goenum:name=ACTIVE
	//goenum:description=Currently active
	//goenum:alias=RUNNING
	StatusActive Status = 1
)

// go:generate goenum generate

s, _ := ParseStatus("RUNNING") // Status
v := int(s)                    // typed
```

Steps:

1. Replace the struct + `EnumBase` with a named basic type and constants.
2. Move names/descriptions/aliases into `//goenum:` directives.
3. Add `//go:generate goenum generate` and run `go generate ./...`.
4. Swap `Statuses.GetByName`/`GetByValue` calls for the generated `Parse…`/
   `…FromValue` functions or `goenum.Set[T]`.
5. Delete the manual `init()` registration.

The two styles can coexist in the same package during migration.

## Performance

`go test -bench .` on this repository (generated vs legacy, same enum):

| Operation | Legacy | Generated |
|---|---|---|
| Parse by name | 26.1 ns | 25.6 ns |
| Parse by alias | 43.8 ns | 30.7 ns |
| Lookup by value | 20.1 ns | 5.3 ns |
| JSON unmarshal | 394 ns / 3 allocs | 380 ns / 2 allocs |

All hot paths are allocation-free and reflection-free. `String()`/`IsValid()`
are a single map lookup (~5 ns); the legacy field read is marginally faster
there, while the generated path wins everywhere a lookup table is involved.

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for your changes
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

Please ensure:
- Code follows Go conventions (`go vet ./...`, `gofmt`)
- Tests pass (`go test ./...`)
- Generated files are regenerated (`go generate ./...`)
- Documentation is updated

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE)
file for details.

## 🙏 Acknowledgments

- Built with Go generics and `go/ast` code generation
- Uses `github.com/stretchr/testify` for testing
- Created with ❤️ by [abdorrahmani](https://github.com/abdorrahmani)
