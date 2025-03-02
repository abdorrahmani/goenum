# GoEnum - A Type-Safe Enum Library for Go

GoEnum is a robust, type-safe enumeration library for Go that leverages generics (Go 1.18+) to provide a clean, efficient, and maintainable way to work with enums. It offers a complete solution for defining enum types, managing sets of enum values, and handling common operations including JSON serialization.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
  - [Basic Operations](#basic-operations)
  - [Enum Set Operations](#enum-set-operations)
  - [JSON Serialization](#json-serialization)
- [How It Works](#how-it-works)
  - [Core Components](#core-components)
  - [Implementation Details](#implementation-details)
- [API Reference](#api-reference)
  - [Enum Interface](#enum-interface)
  - [EnumBase Methods](#enumbase-methods)
  - [EnumSet Methods](#enumset-methods)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Features

- Type-safe enum definitions using Go generics
- Built-in string and integer value support
- Enum set management with lookup capabilities
- Full JSON marshaling/unmarshaling support
- Nil-safe method implementations
- Comprehensive test suite
- Clean, idiomatic Go code
- Extensible design for custom enum types

## Installation

To install GoEnum, run:

```bash
go get github.com/abdorrahmani/goenum
```

Requirements:
- Go 1.18 or higher (for generics support)

## Quick Start

Here's a minimal example to get started:

```go
package main

import (
    "fmt"
    "github.com/abdorrahmani/goenum"
)

type Color struct {
    *goenum.EnumBase
}

var (
    ColorRed   = Color{&goenum.EnumBase{value: 1, name: "RED"}}
    ColorBlue  = Color{&goenum.EnumBase{value: 2, name: "BLUE"}}
    ColorGreen = Color{&goenum.EnumBase{value: 3, name: "GREEN"}}
)

var Colors = goenum.NewEnumSet[Color]()

func init() {
    Colors.Register(ColorRed)
    Colors.Register(ColorBlue)
    Colors.Register(ColorGreen)
}

func main() {
    fmt.Println(ColorRed.String()) // "RED"
    fmt.Println(ColorRed.Value())  // 1
}
```

## Usage Examples

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/abdorrahmani/goenum"
)

type Status struct {
    *goenum.EnumBase
}

var (
    StatusPending = Status{&goenum.EnumBase{value: 0, name: "PENDING"}}
    StatusActive  = Status{&goenum.EnumBase{value: 1, name: "ACTIVE"}}
)

func main() {
    // Basic properties
    fmt.Println(StatusPending.String())  // "PENDING"
    fmt.Println(StatusPending.Value())   // 0
    fmt.Println(StatusPending.IsValid()) // true

    var invalid Status
    fmt.Println(invalid.IsValid()) // false
}
```

### Enum Set Operations

```go
package main

import (
    "fmt"
    "github.com/abdorrahmani/goenum"
)

var Statuses = goenum.NewEnumSet[Status]()

func init() {
    Statuses.Register(StatusPending)
    Statuses.Register(StatusActive)
}

func main() {
    // Lookup by name
    if status, exists := Statuses.GetByName("ACTIVE"); exists {
        fmt.Println(status.Value()) // 1
    }

    // Lookup by value
    if status, exists := Statuses.GetByValue(0); exists {
        fmt.Println(status.String()) // "PENDING"
    }

    // Check existence
    fmt.Println(Statuses.Contains(StatusActive)) // true

    // Get all values
    all := Statuses.Values()
    for _, s := range all {
        fmt.Printf("%s: %d\n", s.String(), s.Value())
    }
}
```

### JSON Serialization

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/abdorrahmani/goenum"
)

type Status struct {
    *goenum.EnumBase
}

func (s Status) MarshalJSON() ([]byte, error) {
    if s.EnumBase == nil {
        return json.Marshal("")
    }
    return s.EnumBase.MarshalJSON()
}

func (s *Status) UnmarshalJSON(data []byte) error {
    if s.EnumBase == nil {
        s.EnumBase = &goenum.EnumBase{}
    }
    return s.EnumBase.UnmarshalJSON(data)
}

func main() {
    // Marshal to JSON
    data, err := json.Marshal(StatusActive)
    if err == nil {
        fmt.Println(string(data)) // "ACTIVE"
    }

    // Unmarshal from JSON
    var status Status
    err = json.Unmarshal([]byte(`"PENDING"`), &status)
    if err == nil {
        fmt.Println(status.String()) // "PENDING"
    }
}
```

## How It Works

### Core Components

1. **Enum Interface**: Defines the contract for all enum types
2. **EnumBase**: Provides a concrete implementation with name and value
3. **EnumSet**: Manages a collection of enum values with lookup methods

### Implementation Details

- **Type Safety**: Uses Go generics to ensure compile-time type checking
- **Nil Safety**: All methods handle nil cases gracefully
- **Storage**: Enum values are stored in a map for efficient lookup
- **JSON**: Custom marshaling/unmarshaling preserves enum names

The library follows a composition pattern where you embed `EnumBase` in your custom type and register instances in an `EnumSet` for management.

## API Reference

### Enum Interface
```go
type Enum interface {
    String() string
    Value() int
    IsValid() bool
}
```

### EnumBase Methods
- `String() string`: Returns the enum name (empty string if nil)
- `Value() int`: Returns the numeric value (0 if nil)
- `IsValid() bool`: Returns true if the enum is initialized
- `MarshalJSON() ([]byte, error)`: JSON marshaling support
- `UnmarshalJSON(data []byte) error`: JSON unmarshaling support

### EnumSet Methods
- `NewEnumSet[T Enum]() *EnumSet[T]`: Creates a new enum set
- `Register(enum T)`: Adds an enum to the set
- `GetByName(name string) (T, bool)`: Retrieves enum by name
- `GetByValue(value int) (T, bool)`: Retrieves enum by value
- `Contains(enum T) bool`: Checks if enum exists in set
- `Values() []T`: Returns all registered enum values

## Best Practices

1. Always register enum values in an `init()` function
2. Use uppercase names for consistency with constants
3. Implement custom JSON methods when embedding in structs
4. Keep enum values unique within a set
5. Use meaningful integer values when applicable

## Testing

The library includes a comprehensive test suite. To run tests:

```bash
go test -v
```

Dependencies:
- `github.com/stretchr/testify` for assertions

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for your changes
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

Please ensure:
- Code follows Go conventions
- Tests pass
- Documentation is updated

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go 1.18+ generics
- Inspired by enum implementations in Java and C#
- Uses `github.com/stretchr/testify` for testing
- Created with ❤️ by [abdorrahmani](https://github.com/abdorrahmani)

