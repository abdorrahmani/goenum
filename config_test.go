package goenum

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// cfgEnum is a dedicated enum type for configuration tests.
type cfgEnum struct {
	*EnumBase
}

var (
	cfgActive   = cfgEnum{NewEnumBase(1, "ACTIVE", "Active", "RUNNING")}
	cfgPending  = cfgEnum{NewEnumBase(2, "PENDING", "Pending", "WAITING")}
	cfgArchived = cfgEnum{NewEnumBase(3, "ARCHIVED", "Archived")}
)

func TestDefaultConfigPreservesV1Behavior(t *testing.T) {
	set := NewEnumSet[cfgEnum]()
	set.Register(cfgActive).Register(cfgPending).Register(cfgArchived)

	assert.Equal(t, DefaultConfig, set.Config())

	t.Run("case-insensitive name lookup", func(t *testing.T) {
		for _, query := range []string{"ACTIVE", "active", "Active"} {
			enum, ok := set.GetByName(query)
			assert.True(t, ok, "GetByName(%q) should match", query)
			assert.Equal(t, cfgActive, enum)
		}
	})

	t.Run("case-insensitive alias lookup", func(t *testing.T) {
		for _, query := range []string{"RUNNING", "running", "Running"} {
			enum, ok := set.GetByName(query)
			assert.True(t, ok, "GetByName(%q) should resolve through alias", query)
			assert.Equal(t, cfgActive, enum)
		}
	})

	t.Run("get by value", func(t *testing.T) {
		enum, ok := set.GetByValue(2)
		assert.True(t, ok)
		assert.Equal(t, cfgPending, enum)

		_, ok = set.GetByValue(99)
		assert.False(t, ok)
	})

	t.Run("aliases remain in metadata", func(t *testing.T) {
		assert.Equal(t, []string{"RUNNING"}, cfgActive.Aliases())
		assert.True(t, cfgActive.HasAlias("running"))
	})

	t.Run("unknown name reports not-found", func(t *testing.T) {
		_, ok := set.GetByName("NOPE")
		assert.False(t, ok)

		_, err := set.Parse("NOPE")
		assert.Error(t, err, "UnknownError is the default for Parse")
	})

	t.Run("default JSON format is name", func(t *testing.T) {
		data, err := json.Marshal(cfgActive)
		assert.NoError(t, err)
		assert.Equal(t, `"ACTIVE"`, string(data))
	})
}

func TestWithJSONFormat(t *testing.T) {
	set := NewEnumSet[cfgEnum](WithJSONFormat(JSONFormatValue))
	set.Register(cfgPending)

	data, err := json.Marshal(cfgPending)
	assert.NoError(t, err)
	assert.Equal(t, `2`, string(data))

	// Instance-level config set after registration still wins.
	cfgPending.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
	data, err = json.Marshal(cfgPending)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"PENDING","value":2,"description":"Pending","aliases":["WAITING"]}`, string(data))
	cfgPending.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatName})
}

func TestWithCaseSensitive(t *testing.T) {
	set := NewEnumSet[cfgEnum](WithCaseSensitive(true))
	set.Register(cfgActive)

	_, ok := set.GetByName("active")
	assert.False(t, ok, "case-sensitive mode must not match 'active'")

	enum, ok := set.GetByName("ACTIVE")
	assert.True(t, ok)
	assert.Equal(t, cfgActive, enum)

	_, err := set.Parse("active")
	assert.Error(t, err)
}

func TestCaseInsensitiveCollisions(t *testing.T) {
	t.Run("name collision", func(t *testing.T) {
		set := NewEnumSet[cfgEnum]()
		set.Register(cfgActive)
		assert.PanicsWithValue(t,
			`ambiguous enum name "active" conflicts with name "ACTIVE" (case-insensitive matching is enabled)`,
			func() {
				set.Register(cfgEnum{NewEnumBase(9, "active", "Lowercase")})
			})
	})

	t.Run("case-sensitive mode allows differing case", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithCaseSensitive(true))
		set.Register(cfgActive)
		lower := cfgEnum{NewEnumBase(9, "active", "Lowercase")}
		assert.NotPanics(t, func() { set.Register(lower) })

		enum, ok := set.GetByName("active")
		assert.True(t, ok)
		assert.Equal(t, lower, enum)
	})
}

func TestAliasConfiguration(t *testing.T) {
	t.Run("aliases disabled", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithAliases(false))
		set.Register(cfgActive).Register(cfgPending)

		enum, ok := set.GetByName("ACTIVE")
		assert.True(t, ok, "canonical name must still resolve")
		assert.Equal(t, cfgActive, enum)

		_, ok = set.GetByName("RUNNING")
		assert.False(t, ok, "aliases must not resolve when disabled")

		_, err := set.Parse("WAITING")
		assert.Error(t, err)

		// Metadata is untouched.
		assert.Equal(t, []string{"RUNNING"}, cfgActive.Aliases())
	})

	t.Run("aliases enabled via option", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithAliases(true))
		set.Register(cfgActive)

		enum, ok := set.GetByName("running")
		assert.True(t, ok)
		assert.Equal(t, cfgActive, enum)
	})
}

func TestAliasCollisions(t *testing.T) {
	t.Run("alias conflicts with other enum name", func(t *testing.T) {
		set := NewEnumSet[cfgEnum]()
		set.Register(cfgActive) // alias RUNNING
		assert.PanicsWithValue(t,
			`ambiguous enum name "RUNNING" conflicts with alias "RUNNING" of enum "ACTIVE"`,
			func() {
				set.Register(cfgEnum{NewEnumBase(9, "RUNNING", "Shadowed")})
			})
	})

	t.Run("alias shared by two enums", func(t *testing.T) {
		set := NewEnumSet[cfgEnum]()
		set.Register(cfgActive)  // alias RUNNING
		set.Register(cfgPending) // alias WAITING
		assert.PanicsWithValue(t,
			`ambiguous alias "WAITING": used by both enum "PENDING" and enum "CLONE"`,
			func() {
				set.Register(cfgEnum{NewEnumBase(9, "CLONE", "Cloned", "WAITING")})
			})
	})

	t.Run("alias equal to own name is allowed", func(t *testing.T) {
		set := NewEnumSet[cfgEnum]()
		self := cfgEnum{NewEnumBase(9, "SELF", "Self alias", "SELF")}
		assert.NotPanics(t, func() { set.Register(self) })

		enum, ok := set.GetByName("self")
		assert.True(t, ok)
		assert.Equal(t, self, enum)
	})

	t.Run("duplicate alias on same enum is allowed", func(t *testing.T) {
		set := NewEnumSet[cfgEnum]()
		dup := cfgEnum{NewEnumBase(9, "DUP", "Duplicate aliases", "X", "X")}
		assert.NotPanics(t, func() { set.Register(dup) })
	})
}

func TestWithUnknownBehavior(t *testing.T) {
	t.Run("UnknownError", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithUnknownBehavior(UnknownError))
		set.Register(cfgActive)

		_, ok := set.GetByName("NOPE")
		assert.False(t, ok)

		_, err := set.Parse("NOPE")
		assert.ErrorContains(t, err, "unknown enum name")

		var target cfgEnum
		target.EnumBase = &EnumBase{}
		set.Register(target) // attach resolver to target
		err = json.Unmarshal([]byte(`"NOPE"`), &target)
		assert.ErrorContains(t, err, "unknown enum name")
	})

	t.Run("UnknownZero", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithUnknownBehavior(UnknownZero))
		set.Register(cfgActive)

		parsed, err := set.Parse("NOPE")
		assert.NoError(t, err)
		assert.False(t, parsed.IsValid(), "unknown name should yield zero enum")

		var target cfgEnum
		target.EnumBase = &EnumBase{}
		set.Register(target)
		err = json.Unmarshal([]byte(`"NOPE"`), &target)
		assert.NoError(t, err)
		assert.False(t, target.IsValid(), "target should be reset to zero enum")
	})

	t.Run("UnknownIgnore", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithUnknownBehavior(UnknownIgnore))
		set.Register(cfgActive)

		parsed, err := set.Parse("NOPE")
		assert.NoError(t, err)
		assert.False(t, parsed.IsValid())

		var target cfgEnum
		target.EnumBase = &EnumBase{}
		set.Register(target)
		err = json.Unmarshal([]byte(`"NOPE"`), &target)
		assert.NoError(t, err)
		assert.False(t, target.IsValid(), "UnknownIgnore leaves zero target unchanged")
	})

	t.Run("known names still resolve", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithUnknownBehavior(UnknownZero))
		set.Register(cfgActive)

		parsed, err := set.Parse("active")
		assert.NoError(t, err)
		assert.Equal(t, cfgActive, parsed)
	})
}

func TestJSONUnmarshalWithResolver(t *testing.T) {
	set := NewEnumSet[cfgEnum]()
	set.Register(cfgActive)

	var target cfgEnum
	target.EnumBase = &EnumBase{}
	set.Register(target) // target now carries the set's resolver

	err := json.Unmarshal([]byte(`"active"`), &target)
	assert.NoError(t, err)
	assert.Equal(t, "ACTIVE", target.String())
	assert.Equal(t, 1, target.Value(), "resolver hydrates the full enum")
	assert.Equal(t, []string{"RUNNING"}, target.Aliases())
}

func TestCombinedConfig(t *testing.T) {
	config := Config{
		JSONFormat:      JSONFormatValue,
		CaseSensitive:   false,
		AllowAliases:    true,
		UnknownBehavior: UnknownError,
	}
	set := NewEnumSet[cfgEnum](WithConfig(config))
	set.Register(cfgActive)

	assert.Equal(t, config, set.Config())

	data, err := json.Marshal(cfgActive)
	assert.NoError(t, err)
	assert.Equal(t, `1`, string(data))

	enum, ok := set.GetByName("running")
	assert.True(t, ok)
	assert.Equal(t, cfgActive, enum)

	_, err = set.Parse("NOPE")
	assert.Error(t, err)
}

func TestConfigNormalization(t *testing.T) {
	t.Run("invalid enum values fall back to defaults", func(t *testing.T) {
		config := Config{
			JSONFormat:      JSONFormat(42),
			AllowAliases:    true,
			UnknownBehavior: UnknownBehavior(42),
		}
		set := NewEnumSet[cfgEnum](WithConfig(config))
		assert.Equal(t, JSONFormatName, set.Config().JSONFormat)
		assert.Equal(t, UnknownError, set.Config().UnknownBehavior)
	})

	t.Run("zero config is strict", func(t *testing.T) {
		set := NewEnumSet[cfgEnum](WithConfig(Config{}))
		assert.False(t, set.Config().AllowAliases)
		assert.Equal(t, JSONFormatName, set.Config().JSONFormat)
	})
}

func TestOptionsCompose(t *testing.T) {
	set := NewEnumSet[cfgEnum](
		WithJSONFormat(JSONFormatFull),
		WithCaseSensitive(true),
		WithAliases(true),
		WithUnknownBehavior(UnknownZero),
	)
	set.Register(cfgActive)

	expected := Config{
		JSONFormat:      JSONFormatFull,
		CaseSensitive:   true,
		AllowAliases:    true,
		UnknownBehavior: UnknownZero,
	}
	assert.Equal(t, expected, set.Config())
}

func TestConfigAccessorReturnsCopy(t *testing.T) {
	set := NewEnumSet[cfgEnum]()
	cfg := set.Config()
	cfg.AllowAliases = false
	assert.True(t, set.Config().AllowAliases, "mutating the returned copy must not affect the set")
}

func TestConfiguredSetConcurrency(t *testing.T) {
	set := NewEnumSet[cfgEnum](WithCaseSensitive(false), WithAliases(true))
	set.Register(cfgActive).Register(cfgPending).Register(cfgArchived)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if enum, ok := set.GetByName("running"); ok {
					assert.Equal(t, cfgActive, enum)
				} else {
					t.Error("concurrent GetByName failed")
					return
				}
				set.GetByValue(2)
				set.Parse("ARCHIVED")
				set.Values()
				set.Names()
				set.Contains(cfgPending)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkGetByName(b *testing.B) {
	set := NewEnumSet[cfgEnum]()
	set.Register(cfgActive).Register(cfgPending).Register(cfgArchived)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.GetByName("ACTIVE")
	}
}

func BenchmarkGetByNameAlias(b *testing.B) {
	set := NewEnumSet[cfgEnum]()
	set.Register(cfgActive).Register(cfgPending).Register(cfgArchived)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.GetByName("RUNNING")
	}
}
