package goenum

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test-local enum types simulating generated code.

type level int

const (
	levelLow level = iota
	levelHigh
)

var levelInfo = Register(EnumInfo[level]{
	Names:   []string{"LOW", "HIGH"},
	Values:  []level{levelLow, levelHigh},
	Descs:   []string{"low level", ""},
	Aliases: [][]string{{"MINOR"}, {"MAJOR", "TOP"}},
})

type code string

const (
	codeOK code = "ok"
)

var codeInfo = Register(EnumInfo[code]{
	Names:  []string{"OK"},
	Values: []code{codeOK},
})

func TestEnumInfoLookup(t *testing.T) {
	v, ok := levelInfo.Lookup("high")
	assert.True(t, ok)
	assert.Equal(t, levelHigh, v)

	v, ok = levelInfo.Lookup("MAJOR")
	assert.True(t, ok)
	assert.Equal(t, levelHigh, v)

	_, ok = levelInfo.Lookup("nope")
	assert.False(t, ok)

	name, ok := levelInfo.Name(levelLow)
	assert.True(t, ok)
	assert.Equal(t, "LOW", name)

	assert.Equal(t, "low level", levelInfo.Description(levelLow))
	assert.Equal(t, "", levelInfo.Description(levelHigh))
	assert.Equal(t, []string{"MAJOR", "TOP"}, levelInfo.AliasesOf(levelHigh))
	assert.True(t, levelInfo.HasAlias(levelHigh, "top"))
	assert.False(t, levelInfo.HasAlias(levelLow, "top"))
}

func TestEnumInfoLookupOpts(t *testing.T) {
	_, ok := levelInfo.LookupOpts("high", true, true)
	assert.False(t, ok, "case-sensitive lookup rejects lowercase")

	_, ok = levelInfo.LookupOpts("major", true, true)
	assert.False(t, ok, "case-sensitive alias lookup rejects lowercase")

	v, ok := levelInfo.LookupOpts("HIGH", true, false)
	assert.True(t, ok)
	assert.Equal(t, levelHigh, v)

	_, ok = levelInfo.LookupOpts("MAJOR", false, false)
	assert.False(t, ok, "aliases disabled")
}

func TestGenericHelpers(t *testing.T) {
	v, err := Parse[level]("minor")
	require.NoError(t, err)
	assert.Equal(t, levelLow, v)

	_, err = Parse[level]("nope")
	assert.ErrorIs(t, err, ErrUnknownEnum)

	assert.Panics(t, func() { MustParse[level]("nope") })
	assert.Equal(t, levelHigh, MustParse[level]("HIGH"))

	assert.Equal(t, "LOW", Name(levelLow))
	assert.Equal(t, "", Name(level(99)))
	assert.True(t, Valid(levelHigh))
	assert.False(t, Valid(level(99)))
	assert.Equal(t, []level{levelLow, levelHigh}, Values[level]())
	assert.Equal(t, []string{"LOW", "HIGH"}, Names[level]())
}

func TestInfoUnregisteredType(t *testing.T) {
	type unregistered int
	_, ok := Info[unregistered]()
	assert.False(t, ok)
	assert.Panics(t, func() { Parse[unregistered]("X") })
}

func TestRegisterPanicsOnMalformedInfo(t *testing.T) {
	type dupLen int
	type dupName int
	type dupVal int
	type aliasClash int

	assert.Panics(t, func() {
		Register(EnumInfo[dupLen]{Names: []string{"A"}, Values: []dupLen{1, 2}})
	})
	assert.Panics(t, func() {
		Register(EnumInfo[dupName]{Names: []string{"A", "A"}, Values: []dupName{1, 2}})
	})
	assert.Panics(t, func() {
		Register(EnumInfo[dupVal]{Names: []string{"A", "B"}, Values: []dupVal{1, 1}})
	})
	assert.Panics(t, func() {
		Register(EnumInfo[aliasClash]{Names: []string{"A", "B"}, Values: []aliasClash{1, 2}, Aliases: [][]string{{"B"}, nil}})
	})
	type foldClash int
	assert.Panics(t, func() {
		Register(EnumInfo[foldClash]{
			Names:   []string{"A", "B"},
			Values:  []foldClash{1, 2},
			Aliases: [][]string{{"RUNNING"}, {"running"}},
		})
	})
	type sameFoldOK int
	assert.NotPanics(t, func() {
		Register(EnumInfo[sameFoldOK]{
			Names:   []string{"A"},
			Values:  []sameFoldOK{1},
			Aliases: [][]string{{"RUNNING", "running"}},
		})
	})
}

func TestStringBackedInfo(t *testing.T) {
	v, ok := codeInfo.Lookup("ok")
	assert.True(t, ok)
	assert.Equal(t, codeOK, v)
	name, ok := codeInfo.Name(codeOK)
	assert.True(t, ok)
	assert.Equal(t, "OK", name)
}

func TestErrUnknownEnumMessage(t *testing.T) {
	_, err := Parse[level]("zzz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"zzz"`)
	assert.True(t, errors.Is(err, ErrUnknownEnum))
}
