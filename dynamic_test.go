package goenum

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicEnumLoading(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "goenum-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test JSON file
	testData := []EnumDefinition{
		{
			Name:        "TEST_A",
			Value:       1,
			Description: "Test enum A",
			Aliases:     []string{"ALPHA"},
		},
		{
			Name:        "TEST_B",
			Value:       2,
			Description: "Test enum B",
			Aliases:     []string{"BETA"},
		},
	}

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	assert.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.json")
	err = os.WriteFile(testFile, jsonData, 0644)
	assert.NoError(t, err)

	// Create validation options that allow duplicates
	options := DefaultValidationOptions()
	options.DuplicateHandling = DuplicateSkip

	t.Run("LoadFromJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromJSON(testFile)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)

		// Verify loaded enums
		enumA, exists := enumSet.GetByName("TEST_A")
		assert.True(t, exists)
		assert.Equal(t, 1, enumA.Value())
		assert.Equal(t, "Test enum A", enumA.Description())
		assert.Equal(t, []string{"ALPHA"}, enumA.Aliases())

		enumB, exists := enumSet.GetByName("TEST_B")
		assert.True(t, exists)
		assert.Equal(t, 2, enumB.Value())
		assert.Equal(t, "Test enum B", enumB.Description())
		assert.Equal(t, []string{"BETA"}, enumB.Aliases())
	})

	t.Run("LoadFromDirectory", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromDirectory(tempDir)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)
		assert.Equal(t, 2, len(enumSet.Values()))
	})

	t.Run("LoadFromMap", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		definitions := map[string]EnumDefinition{
			"TEST_A": testData[0],
			"TEST_B": testData[1],
		}
		err := loader.LoadFromMap(definitions)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)
		assert.Equal(t, 2, len(enumSet.Values()))
	})

	t.Run("LoadFromSlice", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromSlice(testData)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)
		assert.Equal(t, 2, len(enumSet.Values()))
	})

	t.Run("ExportToJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromSlice(testData)
		assert.NoError(t, err)

		exportFile := filepath.Join(tempDir, "export.json")
		err = loader.ExportToJSON(exportFile)
		assert.NoError(t, err)

		// Verify exported file
		data, err := os.ReadFile(exportFile)
		assert.NoError(t, err)

		var exported []EnumDefinition
		err = json.Unmarshal(data, &exported)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(exported))
		assert.Equal(t, "TEST_A", exported[0].Name)
		assert.Equal(t, "TEST_B", exported[1].Name)
	})
}

func TestDynamicEnumLoadingErrors(t *testing.T) {
	// Create options that allow errors for testing error cases
	options := DefaultValidationOptions()
	options.DuplicateHandling = DuplicateError

	t.Run("LoadFromNonExistentFile", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromJSON("nonexistent.json")
		assert.Error(t, err)
	})

	t.Run("LoadFromInvalidJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromReader(&invalidReader{})
		assert.Error(t, err)
	})

	t.Run("LoadFromNonExistentDirectory", func(t *testing.T) {
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromDirectory("nonexistent")
		assert.Error(t, err)
	})
}

// invalidReader is a reader that always returns an error
type invalidReader struct{}

func (r *invalidReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func TestDynamicEnumLoadingEdgeCases(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "goenum-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("empty JSON array", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyNames = true
		options.AllowEmptyValues = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		err := loader.LoadFromSlice([]EnumDefinition{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(loader.GetEnumSet().Values()))
	})

	t.Run("nil values in definition", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyNames = true
		options.AllowEmptyValues = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:        "TEST_NIL",
				Value:       nil,
				Description: "",
				Aliases:     nil,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)

		enum, exists := loader.GetEnumSet().GetByName("TEST_NIL")
		assert.True(t, exists)
		assert.Nil(t, enum.Value())
		assert.Empty(t, enum.Description())
		assert.Empty(t, enum.Aliases())
	})

	t.Run("empty strings in definition", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyNames = true
		options.AllowEmptyValues = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:        "",
				Value:       1,
				Description: "",
				Aliases:     []string{""},
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)

		enum, exists := loader.GetEnumSet().GetByName("")
		assert.True(t, exists)
		assert.Equal(t, 1, enum.Value())
		assert.Empty(t, enum.Description())
		assert.Equal(t, []string{""}, enum.Aliases())
	})

	t.Run("duplicate names in definitions", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateError
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "DUPLICATE",
				Value: 1,
			},
			{
				Name:  "DUPLICATE",
				Value: 2,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate enum found")
	})

	t.Run("duplicate values in definitions", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateError
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "A",
				Value: 1,
			},
			{
				Name:  "B",
				Value: 1,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate enum found")
	})

	t.Run("various value types", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.ValueType = nil                   // Allow any value type
		options.DuplicateHandling = DuplicateSkip // Skip duplicates to avoid errors
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "INT",
				Value: 1,
			},
			{
				Name:  "FLOAT",
				Value: 1.5,
			},
			{
				Name:  "STRING",
				Value: "test",
			},
			{
				Name:  "BOOL",
				Value: true,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)

		// Verify values are preserved correctly
		intEnum, _ := loader.GetEnumSet().GetByName("INT")
		assert.Equal(t, 1, intEnum.Value())

		floatEnum, _ := loader.GetEnumSet().GetByName("FLOAT")
		assert.Equal(t, 1.5, floatEnum.Value())

		stringEnum, _ := loader.GetEnumSet().GetByName("STRING")
		assert.Equal(t, "test", stringEnum.Value())

		boolEnum, _ := loader.GetEnumSet().GetByName("BOOL")
		assert.Equal(t, true, boolEnum.Value())
	})

	t.Run("invalid JSON file content", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
		assert.NoError(t, err)

		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		err = loader.LoadFromJSON(invalidFile)
		assert.Error(t, err)
	})

	t.Run("empty JSON file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(emptyFile, []byte(""), 0644)
		assert.NoError(t, err)

		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		err = loader.LoadFromJSON(emptyFile)
		assert.Error(t, err)
	})

	t.Run("export with empty enum set", func(t *testing.T) {
		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		exportFile := filepath.Join(tempDir, "empty_export.json")
		err := loader.ExportToJSON(exportFile)
		assert.NoError(t, err)

		// Verify exported file contains empty array
		data, err := os.ReadFile(exportFile)
		assert.NoError(t, err)
		assert.Equal(t, "[]", string(data))
	})

	t.Run("load from map with nil values", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyValues = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := map[string]EnumDefinition{
			"TEST_NIL": {
				Name:        "TEST_NIL",
				Value:       nil,
				Description: "",
				Aliases:     nil,
			},
		}
		err := loader.LoadFromMap(definitions)
		assert.NoError(t, err)

		enum, exists := loader.GetEnumSet().GetByName("TEST_NIL")
		assert.True(t, exists)
		assert.Nil(t, enum.Value())
	})

	t.Run("load from directory with no JSON files", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "goenum-empty")
		assert.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		err = loader.LoadFromDirectory(emptyDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no JSON files found")
	})

	t.Run("load from directory with mixed file types", func(t *testing.T) {
		mixedDir, err := os.MkdirTemp("", "goenum-mixed")
		assert.NoError(t, err)
		defer os.RemoveAll(mixedDir)

		// Create a non-JSON file
		nonJsonFile := filepath.Join(mixedDir, "test.txt")
		err = os.WriteFile(nonJsonFile, []byte("not json"), 0644)
		assert.NoError(t, err)

		// Create a valid JSON file
		jsonFile := filepath.Join(mixedDir, "test.json")
		validData := []EnumDefinition{
			{
				Name:  "TEST",
				Value: 1,
			},
		}
		jsonData, err := json.Marshal(validData)
		assert.NoError(t, err)
		err = os.WriteFile(jsonFile, jsonData, 0644)
		assert.NoError(t, err)

		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		err = loader.LoadFromDirectory(mixedDir)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(loader.GetEnumSet().Values()))
	})
}

func TestDynamicEnumValidation(t *testing.T) {
	t.Run("empty name validation", func(t *testing.T) {
		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "",
				Value: 1,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enum name cannot be empty")
	})

	t.Run("empty name allowed", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyNames = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "",
				Value: 1,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)
	})

	t.Run("nil value validation", func(t *testing.T) {
		options := DefaultValidationOptions()
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: nil,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enum value cannot be nil")
	})

	t.Run("nil value allowed", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.AllowEmptyValues = true
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: nil,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)
	})

	t.Run("value type validation", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.ValueType = reflect.TypeOf(0) // Expect int values
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: "string", // Wrong type
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not assignable to expected type")
	})

	t.Run("duplicate handling - error", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateError
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: 1,
			},
			{
				Name:  "TEST",
				Value: 2,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate enum found")
	})

	t.Run("duplicate handling - skip", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateSkip
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: 1,
			},
			{
				Name:  "TEST",
				Value: 2,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)
		enum, exists := loader.GetEnumSet().GetByName("TEST")
		assert.True(t, exists)
		assert.Equal(t, 1, enum.Value()) // First value should be kept
	})

	t.Run("duplicate handling - override", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.DuplicateHandling = DuplicateOverride
		loader := NewDynamicEnumLoader(options)
		definitions := []EnumDefinition{
			{
				Name:  "TEST",
				Value: 1,
			},
			{
				Name:  "TEST",
				Value: 2,
			},
		}
		err := loader.LoadFromSlice(definitions)
		assert.NoError(t, err)
		enum, exists := loader.GetEnumSet().GetByName("TEST")
		assert.True(t, exists)
		assert.Equal(t, 2, enum.Value()) // Second value should override
	})

	t.Run("multiple validations", func(t *testing.T) {
		options := DefaultValidationOptions()
		options.ValueType = reflect.TypeOf("") // Expect string values
		options.AllowEmptyNames = false
		options.AllowEmptyValues = false
		options.DuplicateHandling = DuplicateError
		loader := NewDynamicEnumLoader(options)

		definitions := []EnumDefinition{
			{
				Name:  "", // Empty name
				Value: "test",
			},
			{
				Name:  "TEST",
				Value: nil, // Nil value
			},
			{
				Name:  "TEST2",
				Value: 123, // Wrong type
			},
			{
				Name:  "TEST3",
				Value: "test",
			},
			{
				Name:  "TEST3", // Duplicate
				Value: "test2",
			},
		}

		err := loader.LoadFromSlice(definitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enum name cannot be empty")
	})
}
