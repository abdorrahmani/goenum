package goenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDefaults(t *testing.T) {
	s := NewSet[level]()

	v, ok := s.GetByName("high")
	assert.True(t, ok)
	assert.Equal(t, levelHigh, v)

	v, ok = s.GetByName("MAJOR")
	assert.True(t, ok)
	assert.Equal(t, levelHigh, v)

	name, ok := s.GetByValue(levelLow)
	assert.True(t, ok)
	assert.Equal(t, "LOW", name)

	assert.True(t, s.Contains(levelLow))
	assert.False(t, s.Contains(level(99)))

	assert.Equal(t, []level{levelLow, levelHigh}, s.Values())
	assert.Equal(t, []string{"LOW", "HIGH"}, s.Names())
	assert.Equal(t, "low level", s.Description(levelLow))
	assert.Equal(t, []string{"MINOR"}, s.Aliases(levelLow))
}

func TestSetCaseSensitive(t *testing.T) {
	s := NewSet[level](WithCaseSensitive(true))
	_, ok := s.GetByName("high")
	assert.False(t, ok)
	_, ok = s.GetByName("HIGH")
	assert.True(t, ok)
}

func TestSetAliasesDisabled(t *testing.T) {
	s := NewSet[level](WithAliases(false))
	_, ok := s.GetByName("MAJOR")
	assert.False(t, ok)
	_, ok = s.GetByName("HIGH")
	assert.True(t, ok)
}

func TestSetParseUnknownBehavior(t *testing.T) {
	_, err := NewSet[level]().Parse("nope")
	assert.Error(t, err)

	v, err := NewSet[level](WithUnknownBehavior(UnknownZero)).Parse("nope")
	require.NoError(t, err)
	assert.Equal(t, level(0), v)

	v, err = NewSet[level](WithUnknownBehavior(UnknownIgnore)).Parse("nope")
	require.NoError(t, err)
	assert.Equal(t, level(0), v)

	assert.Panics(t, func() { NewSet[level]().MustParse("nope") })
	assert.Equal(t, levelHigh, NewSet[level]().MustParse("HIGH"))
}

func TestSetConfigCopy(t *testing.T) {
	s := NewSet[level](WithCaseSensitive(true))
	cfg := s.Config()
	assert.True(t, cfg.CaseSensitive)
	cfg.CaseSensitive = false
	assert.True(t, s.Config().CaseSensitive, "Config returns a copy")
}
