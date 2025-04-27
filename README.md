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
  - [Advanced Features](#advanced-features)
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
- Support for both integer and string-based enum values
- Built-in string and value support
- Enum set management with lookup capabilities
- Full JSON marshaling/unmarshaling support
- Nil-safe method implementations
- Comprehensive test suite
- Clean, idiomatic Go code
- Extensible design for custom enum types
- Support for enum descriptions
- Support for enum aliases
- Duplicate value/name validation
- Efficient value-based lookups

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
    ColorRed   = Color{goenum.NewEnumBase(1, "RED", "The color red", "CRIMSON")}
    ColorBlue  = Color{goenum.NewEnumBase(2, "BLUE", "The color blue", "AZURE")}
    ColorGreen = Color{goenum.NewEnumBase(3, "GREEN", "The color green", "EMERALD")}
)

var Colors = goenum.NewEnumSet[Color]()

func init() {
    if err := Colors.Register(ColorRed); err != nil {
        panic(err)
    }
    if err := Colors.Register(ColorBlue); err != nil {
        panic(err)
    }
    if err := Colors.Register(ColorGreen); err != nil {
        panic(err)
    }
}

func main() {
    fmt.Println(ColorRed.String())      // "RED"
    fmt.Println(ColorRed.Value())       // 1
    fmt.Println(ColorRed.Description()) // "The color red"
    fmt.Println(ColorRed.HasAlias("CRIMSON")) // true
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
    StatusPending = Status{goenum.NewEnumBase(0, "PENDING", "Waiting to be processed", "WAITING")}
    StatusActive  = Status{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING")}
)

func main() {
    // Basic properties
    fmt.Println(StatusPending.String())      // "PENDING"
    fmt.Println(StatusPending.Value())       // 0
    fmt.Println(StatusPending.IsValid())     // true
    fmt.Println(StatusPending.Description()) // "Waiting to be processed"
    fmt.Println(StatusPending.HasAlias("WAITING")) // true

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
    if err := Statuses.Register(StatusPending); err != nil {
        panic(err)
    }
    if err := Statuses.Register(StatusActive); err != nil {
        panic(err)
    }
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

    // Lookup by alias
    if status, exists := Statuses.GetByName("WAITING"); exists {
        fmt.Println(status.String()) // "PENDING"
    }

    // Check existence
    fmt.Println(Statuses.Contains(StatusActive)) // true

    // Get all values
    all := Statuses.Values()
    for _, s := range all {
        fmt.Printf("%s: %d (%s)\n", s.String(), s.Value(), s.Description())
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

### Advanced Features

```go
package main

import (
    "fmt"
    "github.com/abdorrahmani/goenum"
)

type Priority struct {
    *goenum.EnumBase
}

var (
    PriorityLow    = Priority{goenum.NewEnumBase("low", "LOW", "Low priority task", "MINOR")}
    PriorityMedium = Priority{goenum.NewEnumBase("medium", "MEDIUM", "Medium priority task", "NORMAL")}
    PriorityHigh   = Priority{goenum.NewEnumBase("high", "HIGH", "High priority task", "URGENT", "CRITICAL")}
)

var Priorities = goenum.NewEnumSet[Priority]()

func init() {
    if err := Priorities.Register(PriorityLow); err != nil {
        panic(err)
    }
    if err := Priorities.Register(PriorityMedium); err != nil {
        panic(err)
    }
    if err := Priorities.Register(PriorityHigh); err != nil {
        panic(err)
    }
}

func main() {
    // String-based enum values
    fmt.Println(PriorityLow.Value()) // "low"

    // Multiple aliases
    fmt.Println(PriorityHigh.HasAlias("URGENT"))  // true
    fmt.Println(PriorityHigh.HasAlias("CRITICAL")) // true
    fmt.Println(PriorityHigh.Aliases()) // ["URGENT", "CRITICAL"]

    // Description support
    fmt.Println(PriorityMedium.Description()) // "Medium priority task"
}
```

## How It Works

### Core Components

1. **Enum Interface**: Defines the contract for all enum types
2. **EnumBase**: Provides a concrete implementation with name, value, description, and aliases
3. **EnumSet**: Manages a collection of enum values with efficient lookup methods

### Implementation Details

- **Type Safety**: Uses Go generics to ensure compile-time type checking
- **Nil Safety**: All methods handle nil cases gracefully
- **Storage**: Enum values are stored in maps for efficient lookup
- **JSON**: Custom marshaling/unmarshaling preserves enum names
- **Validation**: Checks for duplicate names and values
- **Aliases**: Support for multiple string aliases per enum value
- **Descriptions**: Built-in support for enum descriptions

## API Reference

### Enum Interface
```go
type Enum interface {
    String() string
    Value() interface{}
    IsValid() bool
    Description() string
    HasAlias(alias string) bool
    Aliases() []string
}
```

### EnumBase Methods
- `String() string`: Returns the enum name (empty string if nil)
- `Value() interface{}`: Returns the enum value (nil if nil)
- `IsValid() bool`: Returns true if the enum is initialized
- `Description() string`: Returns the enum description
- `HasAlias(alias string) bool`: Checks if the enum has a specific alias
- `Aliases() []string`: Returns all aliases of the enum
- `MarshalJSON() ([]byte, error)`: JSON marshaling support
- `UnmarshalJSON(data []byte) error`: JSON unmarshaling support

### EnumSet Methods
- `NewEnumSet[T Enum]() *EnumSet[T]`: Creates a new enum set
- `Register(enum T) error`: Adds an enum to the set
- `GetByName(name string) (T, bool)`: Retrieves enum by name or alias
- `GetByValue(value interface{}) (T, bool)`: Retrieves enum by value
- `Contains(enum T) bool`: Checks if enum exists in set
- `Values() []T`: Returns all registered enum values

## Best Practices

1. Always register enum values in an `init()` function
2. Use uppercase names for consistency with constants
3. Implement custom JSON methods when embedding in structs
4. Keep enum values unique within a set
5. Use meaningful values and descriptions
6. Use aliases for common alternative names
7. Handle registration errors appropriately
8. Use type-safe enums for better compile-time checking

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

