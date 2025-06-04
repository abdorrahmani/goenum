package goenum

import (
	"fmt"
)

// ExampleReflection demonstrates the usage of reflection features
func ExampleReflection() {
	// Create an enum set for colors
	type Color struct {
		*EnumBase
	}

	// Create color instances
	red := Color{NewEnumBase(1, "RED", "Red color", "PRIMARY")}
	green := Color{NewEnumBase(2, "GREEN", "Green color", "SECONDARY")}
	blue := Color{NewEnumBase(3, "BLUE", "Blue color", "PRIMARY")}

	// Register colors
	colorSet := NewEnumSet[Color]()
	colorSet.Register(red).Register(green).Register(blue)

	// Create a reflection instance
	reflection := NewEnumReflection(colorSet)

	// Get enum fields
	fields, err := reflection.GetEnumFields()
	if err != nil {
		fmt.Printf("Error getting fields: %v\n", err)
		return
	}
	fmt.Println("Enum fields:")
	for _, field := range fields {
		fmt.Printf("  %s: %s\n", field.Name, field.Type)
	}

	// Get enum values
	values, err := reflection.GetEnumValues()
	if err != nil {
		fmt.Printf("Error getting values: %v\n", err)
		return
	}
	fmt.Println("\nEnum values:")
	for _, value := range values {
		fmt.Printf("  %v\n", value)
	}

	// Get enum metadata
	metadata, err := reflection.GetEnumMetadata()
	if err != nil {
		fmt.Printf("Error getting metadata: %v\n", err)
		return
	}
	fmt.Println("\nEnum metadata:")
	fmt.Printf("  Type: %s\n", metadata.Type.Name())
	fmt.Printf("  Package: %s\n", metadata.Type.PkgPath())
	fmt.Printf("  Fields: %d\n", len(metadata.Fields))
	fmt.Printf("  Is composite: %v\n", metadata.IsComposite)

	// Get enum methods
	methods := reflection.GetEnumMethods()
	fmt.Println("\nEnum methods:")
	for _, method := range methods {
		fmt.Printf("  %s\n", method.Name)
	}

	// Get enum interfaces
	interfaces := reflection.GetEnumInterfaces()
	fmt.Println("\nEnum interfaces:")
	for _, iface := range interfaces {
		fmt.Printf("  %s\n", iface.Name())
	}

	// Get enum constants
	constants, err := reflection.GetEnumConstants()
	if err != nil {
		fmt.Printf("Error getting constants: %v\n", err)
		return
	}
	fmt.Println("\nEnum constants:")
	for name, value := range constants {
		fmt.Printf("  %s = %v\n", name, value)
	}
}

// ExampleReflectionWithTags demonstrates the usage of reflection with struct tags
func ExampleReflectionWithTags() {
	// Create an enum with tags
	type TaggedEnum struct {
		*EnumBase `json:"base" xml:"base" yaml:"base"`
		Extra     string `json:"extra" xml:"extra" yaml:"extra"`
	}

	// Create enum instances
	enum1 := TaggedEnum{
		EnumBase: NewEnumBase(1, "ONE", "First enum", "TAG1"),
		Extra:    "extra1",
	}
	enum2 := TaggedEnum{
		EnumBase: NewEnumBase(2, "TWO", "Second enum", "TAG2"),
		Extra:    "extra2",
	}

	// Register enums
	enumSet := NewEnumSet[TaggedEnum]()
	enumSet.Register(enum1).Register(enum2)

	// Create a reflection instance
	reflection := NewEnumReflection(enumSet)

	// Get tags for EnumBase field
	tags, err := reflection.GetEnumTags("EnumBase")
	if err != nil {
		fmt.Printf("Error getting tags: %v\n", err)
		return
	}
	fmt.Println("EnumBase tags:")
	for key, value := range tags {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Get tags for Extra field
	tags, err = reflection.GetEnumTags("Extra")
	if err != nil {
		fmt.Printf("Error getting tags: %v\n", err)
		return
	}
	fmt.Println("\nExtra field tags:")
	for key, value := range tags {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// ExampleReflectionWithComposite demonstrates the usage of reflection with composite enums
func ExampleReflectionWithComposite() {
	// Create a reflection instance for composite enums
	type CompositeEnum struct {
		*EnumBase
	}

	// Create enum instances
	enum1 := CompositeEnum{NewEnumBase(1, "ONE", "First enum", "TAG1")}
	enum2 := CompositeEnum{NewEnumBase(2, "TWO", "Second enum", "TAG2")}

	// Register enums
	enumSet := NewEnumSet[CompositeEnum]()
	enumSet.Register(enum1).Register(enum2)

	// Create a reflection instance
	reflection := NewEnumReflection(enumSet)

	// Get enum metadata
	metadata, err := reflection.GetEnumMetadata()
	if err != nil {
		fmt.Printf("Error getting metadata: %v\n", err)
		return
	}
	fmt.Println("Composite enum metadata:")
	fmt.Printf("  Type: %s\n", metadata.Type.Name())
	fmt.Printf("  Package: %s\n", metadata.Type.PkgPath())
	fmt.Printf("  Fields: %d\n", len(metadata.Fields))
	fmt.Printf("  Is composite: %v\n", metadata.IsComposite)
}
