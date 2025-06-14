# GoEnum - A Type-Safe Enum Library for Go

[![Go Version](https://img.shields.io/badge/Go-1.18%2B-00ADD8?logo=go&logoColor=white)](https://golang.org/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/abdorrahmani/goenum)](https://goreportcard.com/report/github.com/abdorrahmani/goenum)
[![License: MIT](https://img.shields.io/github/license/abdorrahmani/goenum?logo=open-source-initiative&logoColor=white)](https://github.com/abdorrahmani/goenum/blob/main/LICENSE)
[![Coverage](https://img.shields.io/codecov/c/github/abdorrahmani/goenum?logo=codecov)](https://codecov.io/gh/abdorrahmani/goenum)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/abdorrahmani/goenum)
[![GitHub stars](https://img.shields.io/github/stars/abdorrahmani/goenum?style=social)](https://github.com/abdorrahmani/goenum/stargazers)

GoEnum is a powerful, type-safe enumeration library for Go that leverages generics (Go 1.18+) to provide a clean, efficient, and maintainable way to work with enums. It offers a complete solution for defining enum types, managing sets of enum values, and handling common operations including JSON serialization.

## 🌟 Key Features

- **Type Safety**: Leverages Go generics for compile-time type checking
- **Flexible Values**: Support for both integer and string-based enum values
- **Rich Metadata**: Built-in support for descriptions and aliases
- **Efficient Lookups**: Fast value and name-based lookups using maps
- **JSON Support**: Full JSON marshaling/unmarshaling support with multiple formats
- **Nil Safety**: All methods handle nil cases gracefully
- **Validation**: Built-in duplicate value/name checking
- **Extensible**: Easy to extend for custom enum types
- **Well Tested**: Comprehensive test coverage
- **Clean API**: Idiomatic Go code with intuitive interface
- **Composite Enums**: Support for bitwise operations and flag combinations
- **Dynamic Loading**: Load enums from JSON files, maps, or slices

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
  - [Defining Enum Types](#1-defining-enum-types)
  - [Creating Enum Sets](#2-creating-enum-sets)
  - [Working with Aliases](#3-working-with-aliases)
- [Advanced Features](#advanced-features)
  - [JSON Serialization](#1-json-serialization)
  - [String-Based Enums](#2-string-based-enums)
  - [Multiple Aliases](#3-multiple-aliases)
  - [Dynamic Enum Loading](#4-dynamic-enum-loading)
- [Composite Enum Support](#composite-enum-support)
  - [Creating Composite Enums](#creating-composite-enums)
  - [Bitwise Operations](#bitwise-operations)
  - [Type Conversion](#type-conversion)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## 🚀 Installation

```bash
go get github.com/abdorrahmani/goenum
```

**Requirements:**
- Go 1.18 or higher (for generics support)

## 🎯 Quick Start

Here's a minimal example to get started:

```go
package main

import (
    "fmt"
    "github.com/abdorrahmani/goenum"
)

// Define your enum type
type Color struct {
    *goenum.EnumBase
}

// Define enum values
var (
    ColorRed   = Color{goenum.NewEnumBase(1, "RED", "The color red", "CRIMSON")}
    ColorBlue  = Color{goenum.NewEnumBase(2, "BLUE", "The color blue", "AZURE")}
    ColorGreen = Color{goenum.NewEnumBase(3, "GREEN", "The color green", "EMERALD")}
)

// Create an enum set
var Colors = goenum.NewEnumSet[Color]()

func init() {
    // Register enum values
    Colors.Register(ColorRed)
    Colors.Register(ColorBlue)
    Colors.Register(ColorGreen)
}

func main() {
    // Basic usage
    fmt.Println(ColorRed.String())      // "RED"
    fmt.Println(ColorRed.Value())       // 1
    fmt.Println(ColorRed.Description()) // "The color red"
    
    // Lookup by name
    if color, exists := Colors.GetByName("BLUE"); exists {
        fmt.Println(color.Value()) // 2
    }
    
    // Lookup by value
    if color, exists := Colors.GetByValue(3); exists {
        fmt.Println(color.String()) // "GREEN"
    }
}
```

## 📖 Basic Usage

### 1. Defining Enum Types

```go
type Status struct {
    *goenum.EnumBase
}

var (
    StatusPending = Status{goenum.NewEnumBase(0, "PENDING", "Waiting to be processed", "WAITING")}
    StatusActive  = Status{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING")}
    StatusDeleted = Status{goenum.NewEnumBase(2, "DELETED", "The item has been deleted", "REMOVED")}
)
```

### 2. Creating Enum Sets

```go
var Statuses = goenum.NewEnumSet[Status]()

func init() {
    // Using chainable Register method
    Statuses.Register(StatusPending).
        Register(StatusActive).
        Register(StatusDeleted)
}

// Or in a single line
var Colors = goenum.NewEnumSet[Color]().
    Register(ColorRed).
    Register(ColorBlue).
    Register(ColorGreen)

// Usage
if status, exists := Statuses.GetByName("ACTIVE"); exists {
    fmt.Println(status.Value()) // 1
}
```

### 3. Working with Aliases

```go
// Check if an enum has a specific alias
fmt.Println(StatusActive.HasAlias("RUNNING")) // true

// Get all aliases
fmt.Println(StatusActive.Aliases()) // ["RUNNING"]
```

## 🔥 Advanced Features

### 1. JSON Serialization

The library supports three JSON serialization formats:
- `JSONFormatName` (default): Serializes only the enum name
- `JSONFormatValue`: Serializes only the enum value
- `JSONFormatFull`: Serializes a complete struct with name, value, description, and aliases

```go
// Default format (name only)
data, _ := json.Marshal(StatusActive)
fmt.Println(string(data)) // "ACTIVE"

// Value format
StatusActive.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
data, _ = json.Marshal(StatusActive)
fmt.Println(string(data)) // 1

// Full format
StatusActive.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
data, _ = json.Marshal(StatusActive)
fmt.Println(string(data)) // {"name":"ACTIVE","value":1,"description":"Currently active","aliases":["RUNNING"]}

// Unmarshal examples
var status Status
status.EnumBase = &EnumBase{}

// Unmarshal name format
json.Unmarshal([]byte(`"PENDING"`), &status)

// Unmarshal value format
status.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
json.Unmarshal([]byte(`1`), &status)

// Unmarshal full format
status.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
json.Unmarshal([]byte(`{"name":"ACTIVE","value":1,"description":"Currently active","aliases":["RUNNING"]}`), &status)
```

### 2. String-Based Enums

```go
type Priority struct {
    *goenum.EnumBase
}

var (
    PriorityLow    = Priority{goenum.NewEnumBase("low", "LOW", "Low priority task", "MINOR")}
    PriorityMedium = Priority{goenum.NewEnumBase("medium", "MEDIUM", "Medium priority task", "NORMAL")}
    PriorityHigh   = Priority{goenum.NewEnumBase("high", "HIGH", "High priority task", "URGENT", "CRITICAL")}
)
```

### 3. Multiple Aliases

```go
// Define enum with multiple aliases
StatusActive = Status{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING", "LIVE", "ONLINE")}

// Check aliases
fmt.Println(StatusActive.HasAlias("LIVE"))    // true
fmt.Println(StatusActive.HasAlias("ONLINE"))  // true
fmt.Println(StatusActive.Aliases())           // ["RUNNING", "LIVE", "ONLINE"]
```

### 4. Dynamic Enum Loading

The library supports loading enums from various sources:

```go
// Create a loader
loader := goenum.NewDynamicEnumLoader()

// Load from JSON file
err := loader.LoadFromJSON("enums.json")
if err != nil {
    log.Fatal(err)
}

// Load from directory (all JSON files)
err = loader.LoadFromDirectory("enums/")
if err != nil {
    log.Fatal(err)
}

// Load from map
definitions := map[string]goenum.EnumDefinition{
    "TEST_A": {
        Name:        "TEST_A",
        Value:       1,
        Description: "Test enum A",
        Aliases:     []string{"ALPHA"},
    },
}
err = loader.LoadFromMap(definitions)

// Load from slice
definitions := []goenum.EnumDefinition{
    {
        Name:        "TEST_A",
        Value:       1,
        Description: "Test enum A",
        Aliases:     []string{"ALPHA"},
    },
}
err = loader.LoadFromSlice(definitions)

// Export to JSON
err = loader.ExportToJSON("exported_enums.json")
```

Example JSON format for enum definitions:
```json
[
  {
    "name": "TEST_A",
    "value": 1,
    "description": "Test enum A",
    "aliases": ["ALPHA"]
  },
  {
    "name": "TEST_B",
    "value": 2,
    "description": "Test enum B",
    "aliases": ["BETA"]
  }
]
```

## Composite Enum Support

The library supports composite enums that can be combined using bitwise operations. This is particularly useful for flag-based enums where multiple values can be combined.

### Creating Composite Enums

```go
var (
    FlagA = NewCompositeEnumBase(0, "FLAG_A", "First flag")
    FlagB = NewCompositeEnumBase(1, "FLAG_B", "Second flag")
    FlagC = NewCompositeEnumBase(2, "FLAG_C", "Third flag")
)
```

### Bitwise Operations

Composite enums support the following bitwise operations:

- `Or(other CompositeEnum)`: Combines two flags using bitwise OR
- `And(other CompositeEnum)`: Combines two flags using bitwise AND
- `Xor(other CompositeEnum)`: Combines two flags using bitwise XOR
- `Not()`: Inverts the flags using bitwise NOT

Example:
```go
// Combine flags
combined := FlagA.Or(FlagB)  // Results in "FLAG_A|FLAG_B"

// Check if a flag is set
if combined.HasFlag(FlagA) {
    // FlagA is set
}

// Check if flags are empty
if !combined.IsEmpty() {
    // Flags are not empty
}
```

### Type Conversion

The `NewCompositeEnumBase` function accepts various types for the flag value:
- `uint64`: Direct flag value
- `int`: Bit position (value will be 1 << position)
- Other types: Will result in a zero value

## 📚 API Reference

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

### EnumSet Methods

- `NewEnumSet[T Enum]() *EnumSet[T]`: Creates a new enum set
- `Register(enum T) error`: Adds an enum to the set
- `GetByName(name string) (T, bool)`: Retrieves enum by name or alias
- `GetByValue(value interface{}) (T, bool)`: Retrieves enum by value
- `Contains(enum T) bool`: Checks if enum exists in set
- `Values() []T`: Returns all registered enum values
- `Names() []string`: Returns a slice of all enum names
- `Map() map[string]interface{}`: Returns a map of enum names to their values
- `Filter(predicate func(T) bool) []T`: Returns a slice of enums that satisfy the given predicate

## 💡 Best Practices

1. **Initialization**: Always register enum values in an `init()` function
2. **Naming**: Use uppercase names for enum values (e.g., `StatusActive`)
3. **Descriptions**: Provide meaningful descriptions for better documentation
4. **Aliases**: Use aliases for common alternative names
5. **Error Handling**: Check registration errors in `init()`
6. **Type Safety**: Use type-safe enums for better compile-time checking
7. **JSON**: Implement custom JSON methods when embedding in structs
8. **Validation**: Keep enum values unique within a set
9. **Composite Enums**: Use bitwise operations for flag combinations
10. **Dynamic Loading**: Validate enum definitions before loading

## 🤝 Contributing

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

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go 1.18+ generics
- Inspired by enum implementations in Java and C#
- Uses `github.com/stretchr/testify` for testing
- Created with ❤️ by [abdorrahmani](https://github.com/abdorrahmani)
