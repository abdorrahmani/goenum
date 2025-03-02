package goenum

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumBasics(t *testing.T) {
	// Test basic properties
	assert.Equal(t, "PENDING", StatusPending.String())
	assert.Equal(t, 0, StatusPending.Value())
	assert.True(t, StatusPending.IsValid())

	// Test enum set operations
	status, exists := StatusEnumSet.GetByName("ACTIVE")
	assert.True(t, exists)
	assert.Equal(t, 1, status.Value())

	status, exists = StatusEnumSet.GetByValue(2)
	assert.True(t, exists)
	assert.Equal(t, "DELETED", status.String())

	assert.True(t, StatusEnumSet.Contains(StatusActive))
	assert.False(t, StatusEnumSet.Contains(Status{&EnumBase{value: 99, name: "INVALID"}}))
}

func TestJSONMarshaling(t *testing.T) {
	// Test JSON marshaling
	data, err := json.Marshal(StatusActive)
	assert.NoError(t, err)
	assert.Equal(t, `"ACTIVE"`, string(data))

	// Test JSON unmarshaling
	var status Status
	err = json.Unmarshal([]byte(`"PENDING"`), &status)
	assert.NoError(t, err)
	assert.Equal(t, "PENDING", status.String())
}
