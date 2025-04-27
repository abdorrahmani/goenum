package goenum

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnum represents a test enum type
type TestEnum struct {
	*EnumBase
}

var (
	TestEnumA = TestEnum{NewEnumBase(1, "A", "First enum", "ALPHA")}
	TestEnumB = TestEnum{NewEnumBase(2, "B", "Second enum", "BETA")}
	TestEnumC = TestEnum{NewEnumBase(3, "C", "Third enum", "CHARLIE", "THIRD")}
)

var TestEnumSet = NewEnumSet[TestEnum]()

func init() {
	if err := TestEnumSet.Register(TestEnumA); err != nil {
		panic(err)
	}
	if err := TestEnumSet.Register(TestEnumB); err != nil {
		panic(err)
	}
	if err := TestEnumSet.Register(TestEnumC); err != nil {
		panic(err)
	}
}

func TestEnumBasics(t *testing.T) {
	// Test basic properties
	assert.Equal(t, "A", TestEnumA.String())
	assert.Equal(t, 1, TestEnumA.Value())
	assert.True(t, TestEnumA.IsValid())
	assert.Equal(t, "First enum", TestEnumA.Description())

	// Test nil enum
	var nilEnum TestEnum
	assert.Equal(t, "", nilEnum.String())
	assert.Nil(t, nilEnum.Value())
	assert.False(t, nilEnum.IsValid())
	assert.Equal(t, "", nilEnum.Description())
}

func TestEnumAliases(t *testing.T) {
	// Test alias operations
	assert.True(t, TestEnumA.HasAlias("ALPHA"))
	assert.False(t, TestEnumA.HasAlias("BETA"))
	assert.Equal(t, []string{"ALPHA"}, TestEnumA.Aliases())

	// Test multiple aliases
	assert.True(t, TestEnumC.HasAlias("CHARLIE"))
	assert.True(t, TestEnumC.HasAlias("THIRD"))
	assert.ElementsMatch(t, []string{"CHARLIE", "THIRD"}, TestEnumC.Aliases())

	// Test case insensitivity
	assert.True(t, TestEnumA.HasAlias("alpha"))
	assert.True(t, TestEnumA.HasAlias("ALPHA"))
}

func TestEnumSetOperations(t *testing.T) {
	// Test GetByName
	enum, exists := TestEnumSet.GetByName("A")
	assert.True(t, exists)
	assert.Equal(t, TestEnumA, enum)

	// Test GetByName with alias
	enum, exists = TestEnumSet.GetByName("ALPHA")
	assert.True(t, exists)
	assert.Equal(t, TestEnumA, enum)

	// Test GetByValue
	enum, exists = TestEnumSet.GetByValue(2)
	assert.True(t, exists)
	assert.Equal(t, TestEnumB, enum)

	// Test Contains
	assert.True(t, TestEnumSet.Contains(TestEnumA))
	assert.False(t, TestEnumSet.Contains(TestEnum{NewEnumBase(99, "INVALID", "Invalid enum")}))

	// Test Values
	values := TestEnumSet.Values()
	assert.Len(t, values, 3)
	assert.Contains(t, values, TestEnumA)
	assert.Contains(t, values, TestEnumB)
	assert.Contains(t, values, TestEnumC)
}

func TestEnumSetRegistration(t *testing.T) {
	// Test duplicate name registration
	duplicateSet := NewEnumSet[TestEnum]()
	err := duplicateSet.Register(TestEnumA)
	assert.NoError(t, err)

	duplicate := TestEnum{NewEnumBase(99, "A", "Duplicate enum")}
	err = duplicateSet.Register(duplicate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate enum name")

	// Test duplicate value registration
	duplicate = TestEnum{NewEnumBase(1, "DUPLICATE", "Duplicate value")}
	err = duplicateSet.Register(duplicate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate enum value")
}

func TestJSONMarshaling(t *testing.T) {
	// Test JSON marshaling
	data, err := json.Marshal(TestEnumA)
	assert.NoError(t, err)
	assert.Equal(t, `"A"`, string(data))

	// Test JSON marshaling of nil enum
	var nilEnum TestEnum
	data, err = json.Marshal(nilEnum)
	assert.NoError(t, err)
	assert.Equal(t, `""`, string(data))

	// Test JSON unmarshaling
	var enum TestEnum
	enum.EnumBase = &EnumBase{} // Initialize EnumBase
	err = json.Unmarshal([]byte(`"A"`), &enum)
	assert.NoError(t, err)
	assert.Equal(t, TestEnumA.String(), enum.String())

	// Test JSON unmarshaling of empty string
	err = json.Unmarshal([]byte(`""`), &enum)
	assert.NoError(t, err)
	assert.False(t, enum.IsValid())

	// Test JSON unmarshaling into nil EnumBase
	var nilBaseEnum TestEnum
	err = json.Unmarshal([]byte(`"A"`), &nilBaseEnum)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal into nil EnumBase")
}

func TestStringBasedEnum(t *testing.T) {
	// Test string-based enum values
	type StringEnum struct {
		*EnumBase
	}

	var (
		StringEnumA = StringEnum{NewEnumBase("a", "A", "First string enum", "ALPHA")}
		StringEnumB = StringEnum{NewEnumBase("b", "B", "Second string enum", "BETA")}
	)

	stringSet := NewEnumSet[StringEnum]()
	assert.NoError(t, stringSet.Register(StringEnumA))
	assert.NoError(t, stringSet.Register(StringEnumB))

	// Test value operations
	assert.Equal(t, "a", StringEnumA.Value())
	assert.Equal(t, "b", StringEnumB.Value())

	// Test lookup by string value
	enum, exists := stringSet.GetByValue("a")
	assert.True(t, exists)
	assert.Equal(t, StringEnumA, enum)
}

func TestEnumSetEdgeCases(t *testing.T) {
	// Test empty enum set
	emptySet := NewEnumSet[TestEnum]()
	assert.Empty(t, emptySet.Values())

	// Test lookup in empty set
	_, exists := emptySet.GetByName("A")
	assert.False(t, exists)

	// Test lookup with invalid name
	_, exists = TestEnumSet.GetByName("INVALID")
	assert.False(t, exists)

	// Test lookup with invalid value
	_, exists = TestEnumSet.GetByValue(99)
	assert.False(t, exists)
}

func TestEnumDescription(t *testing.T) {
	// Test description operations
	assert.Equal(t, "First enum", TestEnumA.Description())
	assert.Equal(t, "Second enum", TestEnumB.Description())
	assert.Equal(t, "Third enum", TestEnumC.Description())

	// Test empty description
	emptyDesc := TestEnum{NewEnumBase(99, "EMPTY", "")}
	assert.Equal(t, "", emptyDesc.Description())
}
