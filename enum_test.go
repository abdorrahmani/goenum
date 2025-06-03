package goenum

import (
	"encoding/json"
	"strings"
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

func TestJSONSerializationFormats(t *testing.T) {
	t.Run("name format serialization", func(t *testing.T) {
		data, err := json.Marshal(TestEnumA)
		assert.NoError(t, err, "Marshal() should not return error")
		assert.Equal(t, `"A"`, string(data), "Marshal() should return enum name")
	})

	t.Run("value format serialization", func(t *testing.T) {
		TestEnumA.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
		data, err := json.Marshal(TestEnumA)
		assert.NoError(t, err, "Marshal() should not return error")
		assert.Equal(t, `1`, string(data), "Marshal() should return enum value")
	})

	t.Run("full format serialization", func(t *testing.T) {
		TestEnumA.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
		data, err := json.Marshal(TestEnumA)
		assert.NoError(t, err, "Marshal() should not return error")
		expected := `{"name":"A","value":1,"description":"First enum","aliases":["ALPHA"]}`
		assert.JSONEq(t, expected, string(data), "Marshal() should return full enum data")
	})

	t.Run("name format unmarshaling", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{}
		err := json.Unmarshal([]byte(`"A"`), &enum)
		assert.NoError(t, err, "Unmarshal() should not return error")
		assert.Equal(t, "A", enum.String(), "Unmarshal() should set correct name")
	})

	t.Run("value format unmarshaling", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{}
		enum.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
		err := json.Unmarshal([]byte(`1`), &enum)
		assert.NoError(t, err, "Unmarshal() should not return error")
		assert.Equal(t, 1, enum.Value(), "Unmarshal() should set correct value")
	})

	t.Run("full format unmarshaling", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{}
		enum.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
		data := `{"name":"A","value":1,"description":"First enum","aliases":["ALPHA"]}`
		err := json.Unmarshal([]byte(data), &enum)
		assert.NoError(t, err, "Unmarshal() should not return error")
		assert.Equal(t, "A", enum.String(), "Unmarshal() should set correct name")
		assert.Equal(t, 1, enum.Value(), "Unmarshal() should set correct value")
		assert.Equal(t, "First enum", enum.Description(), "Unmarshal() should set correct description")
		assert.Equal(t, []string{"ALPHA"}, enum.Aliases(), "Unmarshal() should set correct aliases")
	})

	t.Run("nil enum handling", func(t *testing.T) {
		var nilEnum TestEnum
		data, err := json.Marshal(nilEnum)
		assert.NoError(t, err, "Marshal() should not return error for nil enum")
		assert.Equal(t, `""`, string(data), "Marshal() should return empty string for nil enum")
	})

	t.Run("invalid json handling", func(t *testing.T) {
		var enum TestEnum
		enum.EnumBase = &EnumBase{}
		err := json.Unmarshal([]byte(`invalid`), &enum)
		assert.Error(t, err, "Unmarshal() should return error for invalid JSON")
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

func TestCompositeEnum(t *testing.T) {
	// Define test flags
	var (
		FlagA = NewCompositeEnumBase(0, "FLAG_A", "First flag")
		FlagB = NewCompositeEnumBase(1, "FLAG_B", "Second flag")
		FlagC = NewCompositeEnumBase(2, "FLAG_C", "Third flag")
	)

	t.Run("basic flag operations", func(t *testing.T) {
		assert.Equal(t, uint64(1), FlagA.Value())
		assert.Equal(t, uint64(2), FlagB.Value())
		assert.Equal(t, uint64(4), FlagC.Value())
	})

	t.Run("bitwise operations", func(t *testing.T) {
		// OR operation
		combined := FlagA.Or(FlagB)
		assert.Equal(t, uint64(3), combined.Value())
		assert.Equal(t, "FLAG_A|FLAG_B", combined.String())

		// AND operation
		andResult := combined.And(FlagA)
		assert.Equal(t, uint64(1), andResult.Value())
		assert.Equal(t, "FLAG_A|FLAG_B&FLAG_A", andResult.String())

		// XOR operation
		xorResult := combined.Xor(FlagA)
		assert.Equal(t, uint64(2), xorResult.Value())
		assert.Equal(t, "FLAG_A|FLAG_B^FLAG_A", xorResult.String())

		// NOT operation
		notResult := FlagA.Not()
		assert.Equal(t, ^uint64(1), notResult.Value())
		assert.Equal(t, "~FLAG_A", notResult.String())
	})

	t.Run("flag checks", func(t *testing.T) {
		combined := FlagA.Or(FlagB)
		assert.True(t, combined.HasFlag(FlagA))
		assert.True(t, combined.HasFlag(FlagB))
		assert.False(t, combined.HasFlag(FlagC))

		empty := &CompositeEnumBase{flags: 0}
		assert.True(t, empty.IsEmpty())
		assert.False(t, combined.IsEmpty())
	})

	t.Run("nil handling", func(t *testing.T) {
		var nilFlag *CompositeEnumBase
		assert.True(t, nilFlag.IsEmpty())
		assert.False(t, nilFlag.HasFlag(FlagA))
		assert.Nil(t, nilFlag.Or(FlagA))
		assert.Nil(t, nilFlag.And(FlagA))
		assert.Nil(t, nilFlag.Xor(FlagA))
		assert.Nil(t, nilFlag.Not())
	})

	t.Run("type conversion", func(t *testing.T) {
		// Test with uint64 value
		flag := NewCompositeEnumBase(uint64(8), "FLAG_D", "Fourth flag")
		assert.Equal(t, uint64(8), flag.Value())

		// Test with int value
		flag = NewCompositeEnumBase(3, "FLAG_E", "Fifth flag")
		assert.Equal(t, uint64(8), flag.Value()) // 1 << 3 = 8

		// Test with invalid value
		flag = NewCompositeEnumBase("invalid", "FLAG_F", "Sixth flag")
		assert.Equal(t, uint64(0), flag.Value())
	})
}

func TestEnumEdgeCases(t *testing.T) {
	t.Run("nil enum base operations", func(t *testing.T) {
		var nilEnum *EnumBase
		assert.Equal(t, "", nilEnum.String())
		assert.Nil(t, nilEnum.Value())
		assert.False(t, nilEnum.IsValid())
		assert.Equal(t, "", nilEnum.Description())
		assert.False(t, nilEnum.HasAlias("any"))
		assert.Nil(t, nilEnum.Aliases())
		assert.Equal(t, DefaultJSONConfig(), nilEnum.GetJSONConfig())
	})

	t.Run("enum with special characters", func(t *testing.T) {
		specialEnum := TestEnum{NewEnumBase(1, "SPECIAL!@#$%^&*()", "Special chars", "ALPHA!@#")}
		assert.Equal(t, "SPECIAL!@#$%^&*()", specialEnum.String())
		assert.True(t, specialEnum.HasAlias("ALPHA!@#"))
	})

	t.Run("enum with unicode characters", func(t *testing.T) {
		unicodeEnum := TestEnum{NewEnumBase(1, "UNICODE_测试_テスト", "Unicode test", "测试", "テスト")}
		assert.Equal(t, "UNICODE_测试_テスト", unicodeEnum.String())
		assert.True(t, unicodeEnum.HasAlias("测试"))
		assert.True(t, unicodeEnum.HasAlias("テスト"))
	})

	t.Run("enum with very long strings", func(t *testing.T) {
		longName := strings.Repeat("A", 1000)
		longDesc := strings.Repeat("B", 1000)
		longAlias := strings.Repeat("C", 1000)

		longEnum := TestEnum{NewEnumBase(1, longName, longDesc, longAlias)}
		assert.Equal(t, longName, longEnum.String())
		assert.Equal(t, longDesc, longEnum.Description())
		assert.True(t, longEnum.HasAlias(longAlias))
	})

	t.Run("enum with whitespace", func(t *testing.T) {
		whitespaceEnum := TestEnum{NewEnumBase(1, "  SPACE  ", "  Description  ", "  ALIAS  ")}
		assert.Equal(t, "  SPACE  ", whitespaceEnum.String())
		assert.Equal(t, "  Description  ", whitespaceEnum.Description())
		assert.True(t, whitespaceEnum.HasAlias("  ALIAS  "))
	})

	t.Run("enum with control characters", func(t *testing.T) {
		controlEnum := TestEnum{NewEnumBase(1, "CONTROL\n\t\r", "Desc\n\t\r", "ALIAS\n\t\r")}
		assert.Equal(t, "CONTROL\n\t\r", controlEnum.String())
		assert.Equal(t, "Desc\n\t\r", controlEnum.Description())
		assert.True(t, controlEnum.HasAlias("ALIAS\n\t\r"))
	})

	t.Run("enum with zero value", func(t *testing.T) {
		zeroEnum := TestEnum{NewEnumBase(0, "ZERO", "Zero value")}
		assert.Equal(t, 0, zeroEnum.Value())
		assert.True(t, zeroEnum.IsValid())
	})

	t.Run("enum with negative value", func(t *testing.T) {
		negativeEnum := TestEnum{NewEnumBase(-1, "NEGATIVE", "Negative value")}
		assert.Equal(t, -1, negativeEnum.Value())
		assert.True(t, negativeEnum.IsValid())
	})

	t.Run("enum with complex value", func(t *testing.T) {
		complexEnum := TestEnum{NewEnumBase(complex(1, 2), "COMPLEX", "Complex value")}
		assert.Equal(t, complex(1, 2), complexEnum.Value())
		assert.True(t, complexEnum.IsValid())
	})

	t.Run("enum with struct value", func(t *testing.T) {
		type TestStruct struct {
			Field1 string
			Field2 int
		}
		structValue := TestStruct{"test", 123}
		structEnum := TestEnum{NewEnumBase(structValue, "STRUCT", "Struct value")}
		assert.Equal(t, structValue, structEnum.Value())
		assert.True(t, structEnum.IsValid())
	})

	t.Run("enum with slice value", func(t *testing.T) {
		sliceValue := []int{1, 2, 3}
		sliceEnum := TestEnum{NewEnumBase(sliceValue, "SLICE", "Slice value")}
		assert.Equal(t, sliceValue, sliceEnum.Value())
		assert.True(t, sliceEnum.IsValid())
	})

	t.Run("enum with map value", func(t *testing.T) {
		mapValue := map[string]int{"a": 1, "b": 2}
		mapEnum := TestEnum{NewEnumBase(mapValue, "MAP", "Map value")}
		assert.Equal(t, mapValue, mapEnum.Value())
		assert.True(t, mapEnum.IsValid())
	})

	t.Run("enum with interface value", func(t *testing.T) {
		var interfaceValue interface{} = "interface value"
		interfaceEnum := TestEnum{NewEnumBase(interfaceValue, "INTERFACE", "Interface value")}
		assert.Equal(t, interfaceValue, interfaceEnum.Value())
		assert.True(t, interfaceEnum.IsValid())
	})

	t.Run("enum with nil value", func(t *testing.T) {
		nilEnum := TestEnum{NewEnumBase(nil, "NIL", "Nil value")}
		assert.Nil(t, nilEnum.Value())
		assert.True(t, nilEnum.IsValid())
	})

	t.Run("enum with empty aliases", func(t *testing.T) {
		emptyAliasesEnum := TestEnum{NewEnumBase(1, "EMPTY_ALIASES", "Empty aliases")}
		assert.Empty(t, emptyAliasesEnum.Aliases())
	})

	t.Run("enum with duplicate aliases", func(t *testing.T) {
		duplicateAliasesEnum := TestEnum{NewEnumBase(1, "DUPLICATE_ALIASES", "Duplicate aliases", "ALIAS", "ALIAS", "ALIAS")}
		assert.Equal(t, []string{"ALIAS", "ALIAS", "ALIAS"}, duplicateAliasesEnum.Aliases())
	})

	t.Run("enum with case-sensitive aliases", func(t *testing.T) {
		caseSensitiveEnum := TestEnum{NewEnumBase(1, "CASE_SENSITIVE", "Case sensitive", "Alias", "ALIAS", "alias")}
		assert.True(t, caseSensitiveEnum.HasAlias("Alias"))
		assert.True(t, caseSensitiveEnum.HasAlias("ALIAS"))
		assert.True(t, caseSensitiveEnum.HasAlias("alias"))
	})

	t.Run("enum with empty alias", func(t *testing.T) {
		emptyAliasEnum := TestEnum{NewEnumBase(1, "EMPTY_ALIAS", "Empty alias", "")}
		assert.True(t, emptyAliasEnum.HasAlias(""))
		assert.Equal(t, []string{""}, emptyAliasEnum.Aliases())
	})

	t.Run("enum with whitespace-only alias", func(t *testing.T) {
		whitespaceAliasEnum := TestEnum{NewEnumBase(1, "WHITESPACE_ALIAS", "Whitespace alias", "   ")}
		assert.True(t, whitespaceAliasEnum.HasAlias("   "))
		assert.Equal(t, []string{"   "}, whitespaceAliasEnum.Aliases())
	})

	t.Run("enum with control characters in alias", func(t *testing.T) {
		controlAliasEnum := TestEnum{NewEnumBase(1, "CONTROL_ALIAS", "Control alias", "\n\t\r")}
		assert.True(t, controlAliasEnum.HasAlias("\n\t\r"))
		assert.Equal(t, []string{"\n\t\r"}, controlAliasEnum.Aliases())
	})

	t.Run("enum with unicode in alias", func(t *testing.T) {
		unicodeAliasEnum := TestEnum{NewEnumBase(1, "UNICODE_ALIAS", "Unicode alias", "测试", "テスト")}
		assert.True(t, unicodeAliasEnum.HasAlias("测试"))
		assert.True(t, unicodeAliasEnum.HasAlias("テスト"))
		assert.Equal(t, []string{"测试", "テスト"}, unicodeAliasEnum.Aliases())
	})

	t.Run("enum with very long alias", func(t *testing.T) {
		longAlias := strings.Repeat("A", 1000)
		longAliasEnum := TestEnum{NewEnumBase(1, "LONG_ALIAS", "Long alias", longAlias)}
		assert.True(t, longAliasEnum.HasAlias(longAlias))
		assert.Equal(t, []string{longAlias}, longAliasEnum.Aliases())
	})
}
