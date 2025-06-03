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
