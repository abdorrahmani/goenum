package goenum

import (
	"fmt"
	"reflect"
	"strings"
)

// EnumMetadata contains reflection-based metadata about an enum type
type EnumMetadata struct {
	// Type information
	Type reflect.Type
	// Field information
	Fields []EnumField
	// Tag information
	Tags map[string]string
	// Value type information
	ValueType reflect.Type
	// IsComposite indicates if this is a composite enum
	IsComposite bool
}

// EnumField represents a field in an enum type
type EnumField struct {
	Name       string
	Type       reflect.Type
	Value      interface{}
	Tags       map[string]string
	IsExported bool
}

// GetEnumMetadata returns reflection-based metadata about an enum type
func GetEnumMetadata[T Enum](enum T) (*EnumMetadata, error) {
	if reflect.ValueOf(enum).IsNil() {
		return nil, fmt.Errorf("cannot get metadata for nil enum")
	}

	// Get the type of the enum
	t := reflect.TypeOf(enum)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	metadata := &EnumMetadata{
		Type:        t,
		Fields:      make([]EnumField, 0),
		Tags:        make(map[string]string),
		ValueType:   reflect.TypeOf(enum.Value()),
		IsComposite: isCompositeEnum(enum),
	}

	// Extract fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := reflect.ValueOf(enum).Field(i).Interface()

		enumField := EnumField{
			Name:       field.Name,
			Type:       field.Type,
			Value:      fieldValue,
			Tags:       make(map[string]string),
			IsExported: field.IsExported(),
		}

		// Extract tags
		for _, tag := range []string{"json", "yaml", "xml", "enum"} {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				enumField.Tags[tag] = tagValue
			}
		}

		metadata.Fields = append(metadata.Fields, enumField)
	}

	// Extract type-level tags
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		// Skip tag extraction for methods as they don't have tags
		metadata.Tags[method.Name] = method.Name
	}

	return metadata, nil
}

// isCompositeEnum checks if an enum is a composite enum
func isCompositeEnum(enum Enum) bool {
	_, ok := enum.(CompositeEnum)
	return ok
}

// GetEnumValueType returns the type of an enum's value
func GetEnumValueType[T Enum](enum T) reflect.Type {
	if reflect.ValueOf(enum).IsNil() {
		return nil
	}
	return reflect.TypeOf(enum.Value())
}

// GetEnumFieldValue returns the value of a specific field in an enum
func GetEnumFieldValue[T Enum](enum T, fieldName string) (interface{}, error) {
	if reflect.ValueOf(enum).IsNil() {
		return nil, fmt.Errorf("cannot get field value from nil enum")
	}

	v := reflect.ValueOf(enum)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found in enum", fieldName)
	}

	return field.Interface(), nil
}

// GetEnumTagValue returns the value of a specific tag on an enum field
func GetEnumTagValue[T Enum](enum T, fieldName, tagName string) (string, error) {
	if reflect.ValueOf(enum).IsNil() {
		return "", fmt.Errorf("cannot get tag value from nil enum")
	}

	t := reflect.TypeOf(enum)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, ok := t.FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("field %s not found in enum", fieldName)
	}

	return field.Tag.Get(tagName), nil
}

// GetEnumFields returns all fields of an enum type
func GetEnumFields[T Enum](enum T) ([]EnumField, error) {
	if reflect.ValueOf(enum).IsNil() {
		return nil, fmt.Errorf("cannot get fields from nil enum")
	}

	t := reflect.TypeOf(enum)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make([]EnumField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := reflect.ValueOf(enum).Field(i).Interface()

		enumField := EnumField{
			Name:       field.Name,
			Type:       field.Type,
			Value:      fieldValue,
			Tags:       make(map[string]string),
			IsExported: field.IsExported(),
		}

		// Extract tags
		for _, tag := range []string{"json", "yaml", "xml", "enum"} {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				enumField.Tags[tag] = tagValue
			}
		}

		fields = append(fields, enumField)
	}

	return fields, nil
}

// GetEnumMethods returns all methods of an enum type
func GetEnumMethods[T Enum](enum T) ([]string, error) {
	if reflect.ValueOf(enum).IsNil() {
		return nil, fmt.Errorf("cannot get methods from nil enum")
	}

	t := reflect.TypeOf(enum)
	methods := make([]string, 0, t.NumMethod())

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methods = append(methods, method.Name)
	}

	return methods, nil
}

// IsEnumType checks if a type implements the Enum interface
func IsEnumType(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Enum)(nil)).Elem())
}

// IsCompositeEnumType checks if a type implements the CompositeEnum interface
func IsCompositeEnumType(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*CompositeEnum)(nil)).Elem())
}

// GetEnumTypeInfo returns information about an enum type
func GetEnumTypeInfo[T Enum](enum T) (map[string]interface{}, error) {
	if reflect.ValueOf(enum).IsNil() {
		return nil, fmt.Errorf("cannot get type info from nil enum")
	}

	t := reflect.TypeOf(enum)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	info := map[string]interface{}{
		"name":         t.Name(),
		"package":      t.PkgPath(),
		"kind":         t.Kind().String(),
		"is_enum":      IsEnumType(t),
		"is_composite": IsCompositeEnumType(t),
		"num_fields":   t.NumField(),
		"num_methods":  t.NumMethod(),
		"value_type":   GetEnumValueType(enum).String(),
		"implements":   make([]string, 0),
	}

	// Check implemented interfaces
	if IsEnumType(t) {
		info["implements"] = append(info["implements"].([]string), "Enum")
	}
	if IsCompositeEnumType(t) {
		info["implements"] = append(info["implements"].([]string), "CompositeEnum")
	}

	return info, nil
}

// EnumReflection provides reflection-based utilities for working with enums
type EnumReflection struct {
	// Type is the reflect.Type of the enum struct
	Type reflect.Type
	// EnumSet is the reflect.Value of the enum set
	EnumSet reflect.Value
}

// NewEnumReflection creates a new EnumReflection instance for the given enum type
func NewEnumReflection[T Enum](enumSet *EnumSet[T]) *EnumReflection {
	var zero T
	return &EnumReflection{
		Type:    reflect.TypeOf(zero),
		EnumSet: reflect.ValueOf(enumSet),
	}
}

// GetEnumFields returns all enum fields from a struct type
func (r *EnumReflection) GetEnumFields() ([]reflect.StructField, error) {
	if r.Type.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %v is not a struct", r.Type)
	}

	var fields []reflect.StructField
	for i := 0; i < r.Type.NumField(); i++ {
		field := r.Type.Field(i)
		if field.Type.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
			fields = append(fields, field)
		}
	}
	return fields, nil
}

// GetEnumValues returns all enum values from a struct type
func (r *EnumReflection) GetEnumValues() ([]Enum, error) {
	if r.Type.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %v is not a struct", r.Type)
	}

	var values []Enum
	for i := 0; i < r.Type.NumField(); i++ {
		field := r.Type.Field(i)
		if field.Type.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
			value := reflect.New(field.Type).Elem()
			if enum, ok := value.Interface().(Enum); ok {
				values = append(values, enum)
			}
		}
	}
	return values, nil
}

// GetEnumSet returns the enum set for a given enum type
func (r *EnumReflection) GetEnumSet() (*EnumSet[Enum], error) {
	if !r.EnumSet.IsValid() {
		return nil, fmt.Errorf("enum set is nil")
	}
	if r.EnumSet.IsNil() {
		return nil, fmt.Errorf("enum set is nil")
	}

	// Get the underlying value
	enumSetValue := r.EnumSet
	if !enumSetValue.IsValid() {
		return nil, fmt.Errorf("invalid enum set value")
	}

	// Create a new EnumSet
	enumSet := NewEnumSet[Enum]()

	// Get the Values() method
	valuesMethod := enumSetValue.MethodByName("Values")
	if !valuesMethod.IsValid() {
		return nil, fmt.Errorf("invalid enum set structure: Values method not found")
	}

	// Call the Values() method to get all enums
	valuesResult := valuesMethod.Call(nil)
	if len(valuesResult) == 0 {
		return nil, fmt.Errorf("invalid enum set structure: Values method returned no results")
	}

	// Get the slice of values
	values := valuesResult[0]
	if values.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid enum set structure: Values method did not return a slice")
	}

	// Register each enum in the new set
	for i := 0; i < values.Len(); i++ {
		value := values.Index(i)
		if enum, ok := value.Interface().(Enum); ok {
			enumSet.Register(enum)
		}
	}

	return enumSet, nil
}

// GetEnumByName uses reflection to find an enum by name
func (r *EnumReflection) GetEnumByName(name string) (Enum, error) {
	if !r.EnumSet.IsValid() {
		return nil, fmt.Errorf("enum set is nil")
	}
	if r.EnumSet.IsNil() {
		return nil, fmt.Errorf("enum set is nil")
	}

	// Get the underlying value
	enumSetValue := r.EnumSet
	if !enumSetValue.IsValid() {
		return nil, fmt.Errorf("invalid enum set value")
	}

	// Get the GetByName method
	getByNameMethod := enumSetValue.MethodByName("GetByName")
	if !getByNameMethod.IsValid() {
		return nil, fmt.Errorf("invalid enum set structure: GetByName method not found")
	}

	// Call GetByName with the name parameter
	result := getByNameMethod.Call([]reflect.Value{reflect.ValueOf(name)})
	if len(result) != 2 {
		return nil, fmt.Errorf("invalid enum set structure: GetByName method returned unexpected results")
	}

	// Check if the enum was found
	if !result[1].Bool() {
		return nil, fmt.Errorf("enum with name %s not found", name)
	}

	// Convert the result to Enum
	if enum, ok := result[0].Interface().(Enum); ok {
		return enum, nil
	}

	return nil, fmt.Errorf("invalid enum type returned by GetByName")
}

// GetEnumByValue uses reflection to find an enum by value
func (r *EnumReflection) GetEnumByValue(value interface{}) (Enum, error) {
	if !r.EnumSet.IsValid() {
		return nil, fmt.Errorf("enum set is nil")
	}
	if r.EnumSet.IsNil() {
		return nil, fmt.Errorf("enum set is nil")
	}

	// Get the underlying value
	enumSetValue := r.EnumSet
	if !enumSetValue.IsValid() {
		return nil, fmt.Errorf("invalid enum set value")
	}

	// Get the GetByValue method
	getByValueMethod := enumSetValue.MethodByName("GetByValue")
	if !getByValueMethod.IsValid() {
		return nil, fmt.Errorf("invalid enum set structure: GetByValue method not found")
	}

	// Call GetByValue with the value parameter
	result := getByValueMethod.Call([]reflect.Value{reflect.ValueOf(value)})
	if len(result) != 2 {
		return nil, fmt.Errorf("invalid enum set structure: GetByValue method returned unexpected results")
	}

	// Check if the enum was found
	if !result[1].Bool() {
		return nil, fmt.Errorf("enum with value %v not found", value)
	}

	// Convert the result to Enum
	if enum, ok := result[0].Interface().(Enum); ok {
		return enum, nil
	}

	return nil, fmt.Errorf("invalid enum type returned by GetByValue")
}

// GetEnumTags returns all tags for a given enum field
func (r *EnumReflection) GetEnumTags(fieldName string) (map[string]string, error) {
	field, ok := r.Type.FieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	tags := make(map[string]string)
	for _, tag := range strings.Split(string(field.Tag), " ") {
		if tag == "" {
			continue
		}
		parts := strings.Split(tag, ":")
		if len(parts) == 2 {
			tags[parts[0]] = strings.Trim(parts[1], "\"")
		}
	}
	return tags, nil
}

// GetEnumMetadata returns metadata about an enum type
func (r *EnumReflection) GetEnumMetadata() (*EnumMetadata, error) {
	fields, err := r.GetEnumFields()
	if err != nil {
		return nil, err
	}

	metadata := &EnumMetadata{
		Type:        r.Type,
		Fields:      make([]EnumField, len(fields)),
		Tags:        make(map[string]string),
		ValueType:   nil, // Will be set if we have any enum fields
		IsComposite: false,
	}

	for i, field := range fields {
		metadata.Fields[i] = EnumField{
			Name:       field.Name,
			Type:       field.Type,
			Value:      nil, // Will be set when we have an instance
			Tags:       make(map[string]string),
			IsExported: field.IsExported(),
		}

		// Extract tags
		for _, tag := range []string{"json", "yaml", "xml", "enum"} {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				metadata.Fields[i].Tags[tag] = tagValue
			}
		}
	}

	return metadata, nil
}

// GetEnumMethods returns all methods available for an enum type
func (r *EnumReflection) GetEnumMethods() []reflect.Method {
	var methods []reflect.Method
	for i := 0; i < r.Type.NumMethod(); i++ {
		method := r.Type.Method(i)
		methods = append(methods, method)
	}
	return methods
}

// GetEnumInterfaces returns all interfaces implemented by the enum type
func (r *EnumReflection) GetEnumInterfaces() []reflect.Type {
	var interfaces []reflect.Type
	for i := 0; i < r.Type.NumMethod(); i++ {
		method := r.Type.Method(i)
		if method.Type.NumIn() > 0 {
			receiver := method.Type.In(0)
			if receiver.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
				interfaces = append(interfaces, receiver)
			}
		}
	}
	return interfaces
}

// GetEnumConstants returns all constant values defined for the enum type
func (r *EnumReflection) GetEnumConstants() (map[string]interface{}, error) {
	if r.Type.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %v is not a struct", r.Type)
	}

	constants := make(map[string]interface{})
	for i := 0; i < r.Type.NumField(); i++ {
		field := r.Type.Field(i)
		if field.Type.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
			value := reflect.New(field.Type).Elem()
			if enum, ok := value.Interface().(Enum); ok {
				constants[field.Name] = enum.Value()
			}
		}
	}
	return constants, nil
}
