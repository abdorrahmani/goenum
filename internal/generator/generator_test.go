package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeAndParse parses src as a single-file package and returns the enums.
func writeAndParse(t *testing.T, src string) []*Enum {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
	enums, err := ParseFiles([]string{path})
	require.NoError(t, err)
	return enums
}

func renderOne(t *testing.T, src string) (string, string) {
	t.Helper()
	enums := writeAndParse(t, src)
	require.Len(t, enums, 1)
	name, out, err := Render(enums[0], "x")
	require.NoError(t, err)
	return name, string(out)
}

func parseErr(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
	_, err := ParseFiles([]string{path})
	require.Error(t, err)
	return err.Error()
}

const simpleSrc = `package x

type Status int

const (
	//goenum:name=PENDING
	//goenum:description=Waiting to be processed
	//goenum:alias=WAITING
	StatusPending Status = iota
	//goenum:name=ACTIVE
	StatusActive
	//goenum:name=DELETED
	StatusDeleted
)
`

func TestSimpleEnumGolden(t *testing.T) {
	name, out := renderOne(t, simpleSrc)
	assert.Equal(t, "status_enum_gen.go", name)
	golden := filepath.Join("testdata", "status_enum_gen.golden")
	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(out), 0o644))
		t.Skip("golden created; inspect and re-run")
	}
	require.NoError(t, err)
	assert.Equal(t, string(want), out)
}

func TestDeterministicOutput(t *testing.T) {
	_, a := renderOne(t, simpleSrc)
	_, b := renderOne(t, simpleSrc)
	assert.Equal(t, a, b)
}

func TestIotaSemantics(t *testing.T) {
	enums := writeAndParse(t, `package x

//goenum:enum
type N int

const (
	A N = iota + 2
	B
	C
)
`)
	require.Len(t, enums, 1)
	vals := []uint64{2, 3, 4}
	for i, v := range enums[0].Values {
		assert.Equal(t, vals[i], v.I64, v.Ident)
	}
}

func TestExplicitAndStringValues(t *testing.T) {
	enums := writeAndParse(t, `package x

//goenum:enum
type Big int64

const (
	BigOne Big = 1000000
	BigTwo Big = -5
)

//goenum:enum
type S string

const (
	SA S = "a"
	SB S = "b"
)
`)
	require.Len(t, enums, 2)
	assert.Equal(t, "Big", enums[0].TypeName)
	assert.Equal(t, "1000000", enums[0].Values[0].Lit)
	assert.Equal(t, "-5", enums[0].Values[1].Lit)
	assert.Equal(t, KindString, enums[1].Kind)
	assert.Equal(t, `"a"`, enums[1].Values[0].Lit)
}

func TestMultipleEnumsOneFile(t *testing.T) {
	enums := writeAndParse(t, `package x

//goenum:enum
type B int

const (
	B0 B = iota
)

//goenum:enum
type A int

const (
	A0 A = iota
)
`)
	require.Len(t, enums, 2)
	assert.Equal(t, []string{"A", "B"}, []string{enums[0].TypeName, enums[1].TypeName}, "sorted by type name")
}

func TestDefaultNames(t *testing.T) {
	enums := writeAndParse(t, `package x

//goenum:enum
type Status int

const (
	StatusPending Status = iota
	StatusActive
)
`)
	require.Len(t, enums, 1)
	assert.Equal(t, []string{"PENDING", "ACTIVE"}, enums[0].Names())
}

func TestOptInRequired(t *testing.T) {
	// A named int with constants but no //goenum directives is not an enum
	// declaration and must be skipped, so `goenum generate ./...` is safe.
	enums := writeAndParse(t, `package x

type Kind int

const (
	KindA Kind = iota
	KindB
)
`)
	assert.Empty(t, enums)
}

func TestFlags(t *testing.T) {
	enums := writeAndParse(t, `package x

// Permission bits.
//
//goenum:flags
type Permission uint64

const (
	//goenum:name=READ
	PermissionRead Permission = 1 << iota
	//goenum:name=WRITE
	PermissionWrite
	//goenum:name=DELETE
	PermissionDelete
)
`)
	require.Len(t, enums, 1)
	require.True(t, enums[0].Flags)
	_, out := renderOne(t, `package x

//goenum:flags
type Permission uint64

const (
	PermissionRead Permission = 1 << iota
	PermissionWrite
)
`)
	assert.Contains(t, out, "func (p Permission) Has(other Permission) bool")
	assert.Contains(t, out, "ParsePermission")
}

func TestIgnoreDirective(t *testing.T) {
	enums := writeAndParse(t, `package x

//goenum:enum
type E int

const (
	E0 E = iota
	//goenum:ignore
	ESentinel E = 99
)
`)
	require.Len(t, enums, 1)
	require.Len(t, enums[0].Values, 1)
}

func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"duplicate name", `package x
type E int
const (
	A E = 1 //goenum:name=X
	B E = 2 //goenum:name=X
)`, `duplicate enum name "X"`},
		{"duplicate value", `package x
//goenum:enum
type E int
const (
	A E = 1
	B E = 1
)`, "duplicate enum value"},
		{"duplicate alias", `package x
type E int
const (
	A E = 1 //goenum:alias=Z
	B E = 2 //goenum:alias=Z
)`, `duplicate enum alias "Z"`},
		{"case-insensitive alias collision", `package x
//goenum:enum
type E int
const (
	A E = 1 //goenum:alias=RUNNING
	B E = 2 //goenum:alias=running
)`, "ambiguous enum alias"},
		{"opted-in enum without constants", `package x
//goenum:enum
type E int`, "declares no constants"},
		{"alias name conflict", `package x
type E int
const (
	A E = 1 //goenum:name=ACTIVE
	B E = 2 //goenum:alias=ACTIVE
)`, "collides with name"},
		{"unknown directive", `package x
type E int
const (
	A E = 1 //goenum:colour=red
)`, "unknown directive"},
		{"directive missing value", `package x
type E int
const (
	A E = 1 //goenum:name
)`, "requires a value"},
		{"unsupported underlying type", `package x
type E float64
const (
	A E = 1 //goenum:name=X
)`, "unsupported underlying type"},
		{"flags not power of two", `package x
//goenum:flags
type F uint64
const (
	A F = 3
)`, "not a power of two"},
		{"flags signed", `package x
//goenum:flags
type F int
const (
	A F = 1
)`, "unsigned integer"},
		{"flags no values", `package x
//goenum:flags
type F uint64`, "declares no constants"},
		{"string enum non-string value", `package x
//goenum:enum
type S string
const (
	A S = 1
)`, "string literals"},
		{"unsupported expression", `package x
//goenum:enum
type E int
const (
	A E = someFunc()
)`, "unsupported const expression"},
		{"position in error", `package x
//goenum:enum
type E int
const (
	A E = 1
	B E = 1
)`, "x.go:6"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := parseErr(t, tc.src)
			assert.Contains(t, msg, tc.want)
		})
	}
}

func TestTrailingDirectives(t *testing.T) {
	enums := writeAndParse(t, `package x

type E int

const (
	A E = 1 //goenum:name=ALPHA
	B E = 2 //goenum:name=BETA
)
`)
	require.Len(t, enums, 1)
	assert.Equal(t, []string{"ALPHA", "BETA"}, enums[0].Names())
}

func TestExpandPathsSkipsGenerated(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(simpleSrc), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status_enum_gen.go"), []byte(simpleSrc), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a_test.go"), []byte(simpleSrc), 0o644))
	paths, err := ExpandPaths([]string{dir})
	require.NoError(t, err)
	require.Len(t, paths, 1)
	assert.True(t, strings.HasSuffix(paths[0], "a.go"))
}

func TestCrossFileEnum(t *testing.T) {
	// Type declared in one file, constants in another: the package is
	// parsed as a whole, and the generated file lands next to the type.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(`package x

type Status int
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte(`package x

const (
	//goenum:name=PENDING
	StatusPending Status = iota
	//goenum:name=ACTIVE
	StatusActive
)
`), 0o644))
	enums, err := ParseFiles([]string{filepath.Join(dir, "a.go")})
	require.NoError(t, err)
	require.Len(t, enums, 1)
	assert.Equal(t, []string{"PENDING", "ACTIVE"}, enums[0].Names())
	assert.Equal(t, filepath.Join(dir, "a.go"), enums[0].Source)
}

func TestPlainConstInExpression(t *testing.T) {
	enums := writeAndParse(t, `package x

const base = 10

//goenum:enum
type E int

const (
	A E = base + 1
	B E = base + 2
)
`)
	require.Len(t, enums, 1)
	assert.Equal(t, uint64(11), enums[0].Values[0].I64)
	assert.Equal(t, uint64(12), enums[0].Values[1].I64)
}

func TestExpandPathsRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(simpleSrc), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "b.go"), []byte(simpleSrc), 0o644))
	paths, err := ExpandPaths([]string{dir + "/..."})
	require.NoError(t, err)
	require.Len(t, paths, 2)
}

func TestGenerateWritesAndNoops(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "x.go"), []byte(simpleSrc), 0o644))

	results, err := Generate(Options{Paths: []string{dir}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].Changed)

	results, err = Generate(Options{Paths: []string{dir}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].Changed, "second run is a no-op")

	// Refuses to clobber a hand-written file at the output path.
	out := results[0].Path
	require.NoError(t, os.WriteFile(out, []byte("package x\n// mine\n"), 0o644))
	_, err = Generate(Options{Paths: []string{dir}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to overwrite")

	results, err = Generate(Options{Paths: []string{dir}, Force: true})
	require.NoError(t, err)
	assert.True(t, results[0].Changed)
}
