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
	// Using chainable Register method
	TestEnumSet.Register(TestEnumA).
		Register(TestEnumB).
		Register(TestEnumC)
}

func TestEnumBasics(t *testing.T) {
	t.Run("valid enum properties", func(t *testing.T) {
		assert.Equal(t, "A", TestEnumA.String(), "String() should return the enum name")
		assert.Equal(t, 1, TestEnumA.Value(), "Value() should return the enum value")
		assert.True(t, TestEnumA.IsValid(), "IsValid() should return true for valid enum")
		assert.Equal(t, "First enum", TestEnumA.Description(), "Description() should return the enum description")
	})

	t.Run("nil enum properties", func(t *testing.T) {
		var nilEnum TestEnum
		assert.Equal(t, "", nilEnum.String(), "String() should return empty string for nil enum")
		assert.Nil(t, nilEnum.Value(), "Value() should return nil for nil enum")
		assert.False(t, nilEnum.IsValid(), "IsValid() should return false for nil enum")
		assert.Equal(t, "", nilEnum.Description(), "Description() should return empty string for nil enum")
	})
}

func TestEnumAliases(t *testing.T) {
	t.Run("single alias operations", func(t *testing.T) {
		assert.True(t, TestEnumA.HasAlias("ALPHA"), "HasAlias() should return true for existing alias")
		assert.False(t, TestEnumA.HasAlias("BETA"), "HasAlias() should return false for non-existing alias")
		assert.Equal(t, []string{"ALPHA"}, TestEnumA.Aliases(), "Aliases() should return all aliases")
	})

	t.Run("multiple aliases operations", func(t *testing.T) {
		assert.True(t, TestEnumC.HasAlias("CHARLIE"), "HasAlias() should return true for first alias")
		assert.True(t, TestEnumC.HasAlias("THIRD"), "HasAlias() should return true for second alias")
		assert.ElementsMatch(t, []string{"CHARLIE", "THIRD"}, TestEnumC.Aliases(), "Aliases() should return all aliases in any order")
	})

	t.Run("case insensitive alias matching", func(t *testing.T) {
		assert.True(t, TestEnumA.HasAlias("alpha"), "HasAlias() should match case-insensitive alias")
		assert.True(t, TestEnumA.HasAlias("ALPHA"), "HasAlias() should match uppercase alias")
		assert.True(t, TestEnumA.HasAlias("Alpha"), "HasAlias() should match mixed-case alias")
	})
}

func TestEnumSetOperations(t *testing.T) {
	t.Run("get by name", func(t *testing.T) {
		enum, exists := TestEnumSet.GetByName("A")
		assert.True(t, exists, "GetByName() should find enum by exact name")
		assert.Equal(t, TestEnumA, enum, "GetByName() should return correct enum")

		enum, exists = TestEnumSet.GetByName("INVALID")
		assert.False(t, exists, "GetByName() should return false for invalid name")
	})

	t.Run("get by alias", func(t *testing.T) {
		enum, exists := TestEnumSet.GetByName("ALPHA")
		assert.True(t, exists, "GetByName() should find enum by alias")
		assert.Equal(t, TestEnumA, enum, "GetByName() should return correct enum for alias")
	})

	t.Run("get by value", func(t *testing.T) {
		enum, exists := TestEnumSet.GetByValue(2)
		assert.True(t, exists, "GetByValue() should find enum by value")
		assert.Equal(t, TestEnumB, enum, "GetByValue() should return correct enum")

		enum, exists = TestEnumSet.GetByValue(99)
		assert.False(t, exists, "GetByValue() should return false for invalid value")
	})

	t.Run("contains check", func(t *testing.T) {
		assert.True(t, TestEnumSet.Contains(TestEnumA), "Contains() should return true for registered enum")
		assert.False(t, TestEnumSet.Contains(TestEnum{NewEnumBase(99, "INVALID", "Invalid enum")}), "Contains() should return false for unregistered enum")
	})

	t.Run("values retrieval", func(t *testing.T) {
		values := TestEnumSet.Values()
		assert.Len(t, values, 3, "Values() should return all registered enums")
		assert.Contains(t, values, TestEnumA, "Values() should contain first enum")
		assert.Contains(t, values, TestEnumB, "Values() should contain second enum")
		assert.Contains(t, values, TestEnumC, "Values() should contain third enum")
	})
}

func TestEnumSetRegistration(t *testing.T) {
	t.Run("duplicate name registration", func(t *testing.T) {
		duplicateSet := NewEnumSet[TestEnum]()
		duplicateSet.Register(TestEnumA)

		duplicate := TestEnum{NewEnumBase(99, "A", "Duplicate enum")}
		assert.Panics(t, func() {
			duplicateSet.Register(duplicate)
		}, "Register() should panic on duplicate name")
	})

	t.Run("duplicate value registration", func(t *testing.T) {
		duplicateSet := NewEnumSet[TestEnum]()
		duplicateSet.Register(TestEnumA)

		duplicate := TestEnum{NewEnumBase(1, "DUPLICATE", "Duplicate value")}
		assert.Panics(t, func() {
			duplicateSet.Register(duplicate)
		}, "Register() should panic on duplicate value")
	})

	t.Run("chainable registration", func(t *testing.T) {
		set := NewEnumSet[TestEnum]()
		result := set.Register(TestEnumA).Register(TestEnumB)
		assert.Equal(t, set, result, "Register() should return the same EnumSet for chaining")
		assert.True(t, set.Contains(TestEnumA), "Chained Register() should register first enum")
		assert.True(t, set.Contains(TestEnumB), "Chained Register() should register second enum")
	})
}

func TestJSONMarshaling(t *testing.T) {
	t.Run("marshal valid enum", func(t *testing.T) {
		data, err := json.Marshal(TestEnumA)
		assert.NoError(t, err, "Marshal() should not return error for valid enum")
		assert.Equal(t, `"A"`, string(data), "Marshal() should return enum name as JSON string")
	})

	t.Run("marshal nil enum", func(t *testing.T) {
		var nilEnum TestEnum
		data, err := json.Marshal(nilEnum)
		assert.NoError(t, err, "Marshal() should not return error for nil enum")
		assert.Equal(t, `""`, string(data), "Marshal() should return empty string for nil enum")
	})

	t.Run("unmarshal valid enum", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{} // Initialize EnumBase
		err := json.Unmarshal([]byte(`"A"`), &enum)
		assert.NoError(t, err, "Unmarshal() should not return error for valid JSON")
		assert.Equal(t, TestEnumA.String(), enum.String(), "Unmarshal() should set correct enum name")
	})

	t.Run("unmarshal empty string", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{} // Initialize EnumBase
		err := json.Unmarshal([]byte(`""`), &enum)
		assert.NoError(t, err, "Unmarshal() should not return error for empty string")
		assert.False(t, enum.IsValid(), "Unmarshal() should result in invalid enum for empty string")
	})

	t.Run("unmarshal into nil EnumBase", func(t *testing.T) {
		var nilBaseEnum TestEnum
		err := json.Unmarshal([]byte(`"A"`), &nilBaseEnum)
		assert.Error(t, err, "Unmarshal() should return error for nil EnumBase")
		assert.Contains(t, err.Error(), "cannot unmarshal into nil EnumBase", "Error message should indicate nil EnumBase")
	})
}

func TestStringBasedEnum(t *testing.T) {
	type StringEnum struct {
		*EnumBase
	}

	var (
		StringEnumA = StringEnum{NewEnumBase("a", "A", "First string enum", "ALPHA")}
		StringEnumB = StringEnum{NewEnumBase("b", "B", "Second string enum", "BETA")}
	)

	t.Run("string value operations", func(t *testing.T) {
		assert.Equal(t, "a", StringEnumA.Value(), "Value() should return string value")
		assert.Equal(t, "b", StringEnumB.Value(), "Value() should return string value")
	})

	t.Run("string value lookup", func(t *testing.T) {
		stringSet := NewEnumSet[StringEnum]()
		stringSet.Register(StringEnumA).Register(StringEnumB)

		enum, exists := stringSet.GetByValue("a")
		assert.True(t, exists, "GetByValue() should find enum by string value")
		assert.Equal(t, StringEnumA, enum, "GetByValue() should return correct enum for string value")

		enum, exists = stringSet.GetByValue("invalid")
		assert.False(t, exists, "GetByValue() should return false for invalid string value")
	})
}

func TestEnumSetEdgeCases(t *testing.T) {
	t.Run("empty enum set", func(t *testing.T) {
		emptySet := NewEnumSet[TestEnum]()
		assert.Empty(t, emptySet.Values(), "Values() should return empty slice for new set")
		assert.False(t, emptySet.Contains(TestEnumA), "Contains() should return false for empty set")
	})

	t.Run("invalid lookups", func(t *testing.T) {
		_, exists := TestEnumSet.GetByName("INVALID")
		assert.False(t, exists, "GetByName() should return false for invalid name")

		_, exists = TestEnumSet.GetByValue(99)
		assert.False(t, exists, "GetByValue() should return false for invalid value")
	})
}

func TestEnumDescription(t *testing.T) {
	t.Run("description operations", func(t *testing.T) {
		assert.Equal(t, "First enum", TestEnumA.Description(), "Description() should return first enum description")
		assert.Equal(t, "Second enum", TestEnumB.Description(), "Description() should return second enum description")
		assert.Equal(t, "Third enum", TestEnumC.Description(), "Description() should return third enum description")
	})

	t.Run("empty description", func(t *testing.T) {
		emptyDesc := TestEnum{NewEnumBase(99, "EMPTY", "")}
		assert.Equal(t, "", emptyDesc.Description(), "Description() should return empty string for empty description")
	})
}

func TestEnumSetUtilityMethods(t *testing.T) {
	t.Run("Names() method", func(t *testing.T) {
		names := TestEnumSet.Names()
		assert.Len(t, names, 3, "Names() should return all enum names")
		assert.Contains(t, names, "A", "Names() should contain first enum name")
		assert.Contains(t, names, "B", "Names() should contain second enum name")
		assert.Contains(t, names, "C", "Names() should contain third enum name")
	})

	t.Run("Map() method", func(t *testing.T) {
		enumMap := TestEnumSet.Map()
		assert.Len(t, enumMap, 3, "Map() should return map with all enums")
		assert.Equal(t, 1, enumMap["A"], "Map() should contain correct value for first enum")
		assert.Equal(t, 2, enumMap["B"], "Map() should contain correct value for second enum")
		assert.Equal(t, 3, enumMap["C"], "Map() should contain correct value for third enum")
	})

	t.Run("Filter() method", func(t *testing.T) {
		// Filter enums with value greater than 1
		filtered := TestEnumSet.Filter(func(e TestEnum) bool {
			return e.Value().(int) > 1
		})
		assert.Len(t, filtered, 2, "Filter() should return correct number of filtered enums")
		assert.Contains(t, filtered, TestEnumB, "Filter() should contain second enum")
		assert.Contains(t, filtered, TestEnumC, "Filter() should contain third enum")
		assert.NotContains(t, filtered, TestEnumA, "Filter() should not contain first enum")

		// Filter enums with specific description
		filtered = TestEnumSet.Filter(func(e TestEnum) bool {
			return e.Description() == "First enum"
		})
		assert.Len(t, filtered, 1, "Filter() should return single enum with matching description")
		assert.Contains(t, filtered, TestEnumA, "Filter() should contain enum with matching description")
	})
}
