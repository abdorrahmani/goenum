package goenum

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	t.Run("LoadFromJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
		loader := NewDynamicEnumLoader()
		err := loader.LoadFromDirectory(tempDir)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)
		assert.Equal(t, 2, len(enumSet.Values()))
	})

	t.Run("LoadFromMap", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
		loader := NewDynamicEnumLoader()
		err := loader.LoadFromSlice(testData)
		assert.NoError(t, err)

		enumSet := loader.GetEnumSet()
		assert.NotNil(t, enumSet)
		assert.Equal(t, 2, len(enumSet.Values()))
	})

	t.Run("ExportToJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
	t.Run("LoadFromNonExistentFile", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
		err := loader.LoadFromJSON("nonexistent.json")
		assert.Error(t, err)
	})

	t.Run("LoadFromInvalidJSON", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
		err := loader.LoadFromReader(&invalidReader{})
		assert.Error(t, err)
	})

	t.Run("LoadFromNonExistentDirectory", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
		loader := NewDynamicEnumLoader()
		err := loader.LoadFromSlice([]EnumDefinition{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(loader.GetEnumSet().Values()))
	})

	t.Run("nil values in definition", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
		loader := NewDynamicEnumLoader()
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
		loader := NewDynamicEnumLoader()
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
		assert.Panics(t, func() {
			loader.LoadFromSlice(definitions)
		})
	})

	t.Run("duplicate values in definitions", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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
		assert.Panics(t, func() {
			loader.LoadFromSlice(definitions)
		})
	})

	t.Run("various value types", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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

		loader := NewDynamicEnumLoader()
		err = loader.LoadFromJSON(invalidFile)
		assert.Error(t, err)
	})

	t.Run("empty JSON file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(emptyFile, []byte(""), 0644)
		assert.NoError(t, err)

		loader := NewDynamicEnumLoader()
		err = loader.LoadFromJSON(emptyFile)
		assert.Error(t, err)
	})

	t.Run("export with empty enum set", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
		exportFile := filepath.Join(tempDir, "empty_export.json")
		err := loader.ExportToJSON(exportFile)
		assert.NoError(t, err)

		// Verify exported file contains empty array
		data, err := os.ReadFile(exportFile)
		assert.NoError(t, err)
		assert.Equal(t, "[]", string(data))
	})

	t.Run("load from map with nil values", func(t *testing.T) {
		loader := NewDynamicEnumLoader()
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

		loader := NewDynamicEnumLoader()
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

		loader := NewDynamicEnumLoader()
		err = loader.LoadFromDirectory(mixedDir)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(loader.GetEnumSet().Values()))
	})
}
