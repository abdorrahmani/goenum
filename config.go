package goenum

// UnknownBehavior defines how lookups, parsing and JSON unmarshaling react to
// names that are not registered in an EnumSet.
type UnknownBehavior int

const (
	// UnknownError returns an error from Parse and UnmarshalJSON when the
	// name cannot be resolved. Set lookups that return (T, bool) report
	// ok=false. This is the default and matches v1 behavior for lookups.
	UnknownError UnknownBehavior = iota
	// UnknownZero resolves unknown names to the zero enum without an error.
	// In UnmarshalJSON the target is reset to a zero enum.
	UnknownZero
	// UnknownIgnore behaves like UnknownZero for Parse. In UnmarshalJSON the
	// target is left unchanged instead of being reset.
	UnknownIgnore
)

// Config centralizes the behavior of an EnumSet. The zero value is the
// strictest configuration (no aliases, case-sensitive-free lookups). Use
// DefaultConfig, individual options, or WithConfig to get v1-compatible
// behavior.
type Config struct {
	// JSONFormat selects how registered enums serialize to JSON
	// (JSONFormatName, JSONFormatValue or JSONFormatFull).
	JSONFormat JSONFormat
	// CaseSensitive controls name and alias matching. When true, "ACTIVE"
	// and "active" are different keys. When false, matching is
	// case-insensitive.
	CaseSensitive bool
	// AllowAliases controls whether aliases are accepted during lookups,
	// parsing and JSON name unmarshaling. Aliases stay part of the enum
	// metadata either way.
	AllowAliases bool
	// UnknownBehavior selects the reaction to unknown names.
	UnknownBehavior UnknownBehavior
}

// DefaultConfig preserves the v1 GoEnum behavior: name-format JSON,
// case-insensitive lookups, aliases enabled, unknown names reported as
// not-found / errors. Treat it as read-only; mutating it affects every
// EnumSet created without explicit configuration.
var DefaultConfig = Config{
	JSONFormat:      JSONFormatName,
	CaseSensitive:   false,
	AllowAliases:    true,
	UnknownBehavior: UnknownError,
}

// normalized replaces out-of-range enum values with their defaults so the
// zero value of JSONFormat/UnknownBehavior stays meaningful.
func (c Config) normalized() Config {
	switch c.JSONFormat {
	case JSONFormatName, JSONFormatValue, JSONFormatFull:
	default:
		c.JSONFormat = JSONFormatName
	}
	switch c.UnknownBehavior {
	case UnknownError, UnknownZero, UnknownIgnore:
	default:
		c.UnknownBehavior = UnknownError
	}
	return c
}

// EnumSetOption configures an EnumSet at construction time. Options are meant
// to be passed to NewEnumSet; configuration is immutable afterwards.
type EnumSetOption func(*Config)

// WithConfig applies a complete configuration to the EnumSet. The zero value
// of Config disables aliases; use DefaultConfig for v1-compatible defaults.
func WithConfig(config Config) EnumSetOption {
	return func(c *Config) {
		*c = config
	}
}

// WithJSONFormat sets the JSON serialization format for enums registered in
// the set.
func WithJSONFormat(format JSONFormat) EnumSetOption {
	return func(c *Config) {
		c.JSONFormat = format
	}
}

// WithCaseSensitive toggles case-sensitive name and alias matching.
func WithCaseSensitive(caseSensitive bool) EnumSetOption {
	return func(c *Config) {
		c.CaseSensitive = caseSensitive
	}
}

// WithAliases toggles alias resolution during lookups, parsing and JSON name
// unmarshaling.
func WithAliases(allow bool) EnumSetOption {
	return func(c *Config) {
		c.AllowAliases = allow
	}
}

// WithUnknownBehavior sets how unknown names are handled.
func WithUnknownBehavior(behavior UnknownBehavior) EnumSetOption {
	return func(c *Config) {
		c.UnknownBehavior = behavior
	}
}
