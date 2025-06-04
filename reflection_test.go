package goenum

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ReflectionTestEnum represents a test enum type with reflection support
type ReflectionTestEnum struct {
	*EnumBase
}

var (
	ReflectionTestEnumA = ReflectionTestEnum{NewEnumBase(1, "A", "First enum", "ALPHA")}
	ReflectionTestEnumB = ReflectionTestEnum{NewEnumBase(2, "B", "Second enum", "BETA")}
	ReflectionTestEnumC = ReflectionTestEnum{NewEnumBase(3, "C", "Third enum", "CHARLIE", "THIRD")}
)

var ReflectionTestEnumSet = NewEnumSet[ReflectionTestEnum]()

func init() {
	ReflectionTestEnumSet.Register(ReflectionTestEnumA).
		Register(ReflectionTestEnumB).
		Register(ReflectionTestEnumC)
}

func TestEnumReflection(t *testing.T) {
	t.Run("NewEnumReflection", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		assert.NotNil(t, reflection)
		assert.Equal(t, reflect.TypeOf(ReflectionTestEnum{}), reflection.Type)
		assert.Equal(t, reflect.ValueOf(ReflectionTestEnumSet), reflection.EnumSet)
	})

	t.Run("GetEnumFields", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		fields, err := reflection.GetEnumFields()
		assert.NoError(t, err)
		assert.NotEmpty(t, fields)
	})

	t.Run("GetEnumValues", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		values, err := reflection.GetEnumValues()
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
	})

	t.Run("GetEnumSet", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		enumSet, err := reflection.GetEnumSet()
		assert.NoError(t, err)
		assert.NotNil(t, enumSet)
	})

	t.Run("GetEnumByName", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		enum, err := reflection.GetEnumByName("A")
		assert.NoError(t, err)
		assert.NotNil(t, enum)
		assert.Equal(t, "A", enum.String())
	})

	t.Run("GetEnumByValue", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		enum, err := reflection.GetEnumByValue(1)
		assert.NoError(t, err)
		assert.NotNil(t, enum)
		assert.Equal(t, 1, enum.Value())
	})

	t.Run("GetEnumTags", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		tags, err := reflection.GetEnumTags("EnumBase")
		assert.NoError(t, err)
		assert.NotNil(t, tags)
	})

	t.Run("GetEnumMetadata", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		metadata, err := reflection.GetEnumMetadata()
		assert.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.Equal(t, "ReflectionTestEnum", metadata.Type.Name())
		assert.NotEmpty(t, metadata.Fields)
	})

	t.Run("GetEnumMethods", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		methods := reflection.GetEnumMethods()
		assert.NotEmpty(t, methods)
	})

	t.Run("GetEnumInterfaces", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		interfaces := reflection.GetEnumInterfaces()
		assert.NotEmpty(t, interfaces)
	})

	t.Run("GetEnumConstants", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		constants, err := reflection.GetEnumConstants()
		assert.NoError(t, err)
		assert.NotEmpty(t, constants)
	})
}

func TestEnumReflectionErrors(t *testing.T) {
	t.Run("GetEnumFields with non-struct type", func(t *testing.T) {
		reflection := &EnumReflection{
			Type: reflect.TypeOf(1),
		}
		_, err := reflection.GetEnumFields()
		assert.Error(t, err)
	})

	t.Run("GetEnumValues with non-struct type", func(t *testing.T) {
		reflection := &EnumReflection{
			Type: reflect.TypeOf(1),
		}
		_, err := reflection.GetEnumValues()
		assert.Error(t, err)
	})

	t.Run("GetEnumSet with nil enum set", func(t *testing.T) {
		reflection := &EnumReflection{
			Type:    reflect.TypeOf(ReflectionTestEnum{}),
			EnumSet: reflect.ValueOf(nil),
		}
		_, err := reflection.GetEnumSet()
		assert.Error(t, err)
	})

	t.Run("GetEnumByName with non-existent name", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		_, err := reflection.GetEnumByName("NON_EXISTENT")
		assert.Error(t, err)
	})

	t.Run("GetEnumByValue with non-existent value", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		_, err := reflection.GetEnumByValue(999)
		assert.Error(t, err)
	})

	t.Run("GetEnumTags with non-existent field", func(t *testing.T) {
		reflection := NewEnumReflection(ReflectionTestEnumSet)
		_, err := reflection.GetEnumTags("NON_EXISTENT")
		assert.Error(t, err)
	})

	t.Run("GetEnumConstants with non-struct type", func(t *testing.T) {
		reflection := &EnumReflection{
			Type: reflect.TypeOf(1),
		}
		_, err := reflection.GetEnumConstants()
		assert.Error(t, err)
	})
}

func TestEnumReflectionEdgeCases(t *testing.T) {
	t.Run("GetEnumMetadata with empty struct", func(t *testing.T) {
		type EmptyEnum struct{}
		reflection := &EnumReflection{
			Type: reflect.TypeOf(EmptyEnum{}),
		}
		metadata, err := reflection.GetEnumMetadata()
		assert.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.Empty(t, metadata.Fields)
	})

	t.Run("GetEnumMethods with no methods", func(t *testing.T) {
		type NoMethodEnum struct{}
		reflection := &EnumReflection{
			Type: reflect.TypeOf(NoMethodEnum{}),
		}
		methods := reflection.GetEnumMethods()
		assert.Empty(t, methods)
	})

	t.Run("GetEnumInterfaces with no interfaces", func(t *testing.T) {
		type NoInterfaceEnum struct{}
		reflection := &EnumReflection{
			Type: reflect.TypeOf(NoInterfaceEnum{}),
		}
		interfaces := reflection.GetEnumInterfaces()
		assert.Empty(t, interfaces)
	})

	t.Run("GetEnumConstants with no constants", func(t *testing.T) {
		type NoConstantEnum struct{}
		reflection := &EnumReflection{
			Type: reflect.TypeOf(NoConstantEnum{}),
		}
		constants, err := reflection.GetEnumConstants()
		assert.NoError(t, err)
		assert.Empty(t, constants)
	})
}
