package status_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	goenum "github.com/abdorrahmani/goenum"
	. "github.com/abdorrahmani/goenum/example/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedScalarAPI(t *testing.T) {
	assert.Equal(t, "ACTIVE", StatusActive.String())
	assert.Equal(t, "Status(9)", Status(9).String())
	assert.True(t, StatusActive.IsValid())
	assert.False(t, Status(9).IsValid())
	assert.Equal(t, "Currently active", StatusActive.Description())
	assert.Equal(t, "", Status(9).Description())
	assert.Equal(t, []string{"RUNNING"}, StatusActive.Aliases())
	assert.True(t, StatusActive.HasAlias("running"))
	assert.False(t, StatusPending.HasAlias("running"))

	v, err := ParseStatus("RUNNING")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, v)

	_, err = ParseStatus("nope")
	assert.ErrorIs(t, err, goenum.ErrUnknownEnum)

	assert.Panics(t, func() { MustParseStatus("nope") })
	assert.Equal(t, StatusDeleted, MustParseStatus("DELETED"))

	s, ok := StatusFromValue(1)
	assert.True(t, ok)
	assert.Equal(t, StatusActive, s)
	_, ok = StatusFromValue(9)
	assert.False(t, ok)
}

func TestGeneratedStringBackedAPI(t *testing.T) {
	assert.Equal(t, "HIGH", PriorityHigh.String())
	assert.Equal(t, "high", string(PriorityHigh))
	v, err := ParsePriority("URGENT")
	require.NoError(t, err)
	assert.Equal(t, PriorityHigh, v)
	p, ok := PriorityFromValue("low")
	assert.True(t, ok)
	assert.Equal(t, PriorityLow, p)
}

func TestGeneratedJSON(t *testing.T) {
	data, err := json.Marshal(StatusActive)
	require.NoError(t, err)
	assert.Equal(t, `"ACTIVE"`, string(data))

	var s Status
	require.NoError(t, json.Unmarshal([]byte(`"REMOVED"`), &s))
	assert.Equal(t, StatusDeleted, s)

	err = json.Unmarshal([]byte(`"nope"`), &s)
	assert.Error(t, err)

	_, err = json.Marshal(Status(9))
	assert.Error(t, err, "undeclared values must not marshal silently")

	// In a struct.
	type payload struct {
		Status   Status   `json:"status"`
		Priority Priority `json:"priority"`
	}
	data, err = json.Marshal(payload{StatusActive, PriorityHigh})
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ACTIVE","priority":"HIGH"}`, string(data))

	var back payload
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, payload{StatusActive, PriorityHigh}, back)
}

func TestGeneratedText(t *testing.T) {
	text, err := StatusActive.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", string(text))

	var s Status
	require.NoError(t, s.UnmarshalText([]byte("waiting")))
	assert.Equal(t, StatusPending, s)
}

func TestGeneratedSQL(t *testing.T) {
	v, err := StatusActive.Value()
	require.NoError(t, err)
	assert.Equal(t, int64(1), v)

	var s Status
	require.NoError(t, s.Scan(int64(2)))
	assert.Equal(t, StatusDeleted, s)
	require.NoError(t, s.Scan("RUNNING"))
	assert.Equal(t, StatusActive, s)
	require.NoError(t, s.Scan([]byte("ACTIVE")))
	assert.Equal(t, StatusActive, s)
	require.NoError(t, s.Scan(nil))
	assert.Equal(t, Status(0), s)
	assert.Error(t, s.Scan(int64(9)))
	assert.Error(t, s.Scan(3.5))

	// String-backed stores its underlying string.
	pv, err := PriorityHigh.Value()
	require.NoError(t, err)
	assert.Equal(t, "high", pv)

	// Usable through database/sql interfaces.
	var valuer interface{ Value() (any, error) } = StatusActive
	_ = valuer
	var scanner sql.Scanner = &s
	assert.NotNil(t, scanner)
}

func TestGeneratedFlags(t *testing.T) {
	p := PermissionRead.Add(PermissionWrite)
	assert.True(t, p.Has(PermissionRead))
	assert.True(t, p.Has(PermissionWrite))
	assert.False(t, p.Has(PermissionDelete))
	assert.Equal(t, "READ|WRITE", p.String())

	p = p.Remove(PermissionRead)
	assert.Equal(t, PermissionWrite, p)
	assert.False(t, p.IsEmpty())
	assert.True(t, Permission(0).IsEmpty())

	assert.True(t, PermissionRead.IsValid())
	assert.False(t, Permission(8).IsValid())
	assert.True(t, Permission(0).IsValid())

	v, err := ParsePermission("READ|DELETE")
	require.NoError(t, err)
	assert.Equal(t, PermissionRead.Add(PermissionDelete), v)

	zero, err := ParsePermission("0")
	require.NoError(t, err)
	assert.Equal(t, Permission(0), zero)

	_, err = ParsePermission("READ|NOPE")
	assert.ErrorIs(t, err, goenum.ErrUnknownEnum)

	data, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Equal(t, `"WRITE"`, string(data))

	var back Permission
	require.NoError(t, json.Unmarshal([]byte(`"READ|WRITE"`), &back))
	assert.Equal(t, PermissionRead.Add(PermissionWrite), back)
}

func TestGenericHelpersOnGeneratedTypes(t *testing.T) {
	s, err := goenum.Parse[Status]("active")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, s)

	assert.Equal(t, "PENDING", goenum.Name(StatusPending))
	assert.True(t, goenum.Valid(StatusDeleted))
	assert.False(t, goenum.Valid(Status(9)))
	assert.Equal(t, []Status{StatusPending, StatusActive, StatusDeleted}, goenum.Values[Status]())
	assert.Equal(t, []string{"PENDING", "ACTIVE", "DELETED"}, goenum.Names[Status]())
}

func TestSetOnGeneratedTypes(t *testing.T) {
	statuses := goenum.NewSet[Status]()
	v, ok := statuses.GetByName("RUNNING")
	assert.True(t, ok)
	assert.Equal(t, StatusActive, v)

	name, ok := statuses.GetByValue(StatusDeleted)
	assert.True(t, ok)
	assert.Equal(t, "DELETED", name)

	// Compile-time type safety: this would not compile with the legacy
	// EnumSet.GetByValue(interface{}):
	//   statuses.GetByValue("wrong type")
	_ = fmt.Sprint(name)
}

func TestLegacyInteropUnaffected(t *testing.T) {
	// The legacy EnumBase API still works alongside generated enums.
	base := goenum.NewEnumBase(1, "X", "legacy")
	assert.Equal(t, "X", base.String())
	assert.Equal(t, 1, base.Value())
	assert.True(t, errors.Is(goenum.ErrUnknownEnum, goenum.ErrUnknownEnum))
}

// TestGeneratedStringBackedFullAPI exercises the string-backed generated
// methods left uncovered by TestGeneratedStringBackedAPI: metadata accessors,
// text/SQL marshaling, MustParse and undeclared-value rendering.
func TestGeneratedStringBackedFullAPI(t *testing.T) {
	assert.Equal(t, "Low priority task", PriorityLow.Description())
	assert.Equal(t, []string{"URGENT"}, PriorityHigh.Aliases())
	assert.Empty(t, PriorityLow.Aliases())
	assert.True(t, PriorityHigh.HasAlias("urgent"))
	assert.False(t, PriorityLow.HasAlias("urgent"))

	assert.Equal(t, PriorityHigh, MustParsePriority("HIGH"))
	assert.Panics(t, func() { MustParsePriority("nope") })

	// Undeclared value renders quoted and does not resolve.
	assert.Equal(t, `Priority("nope")`, Priority("nope").String())
	assert.False(t, Priority("nope").IsValid())
	_, ok := PriorityFromValue("nope")
	assert.False(t, ok)

	// Text marshaling round-trips names and accepts aliases.
	text, err := PriorityHigh.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "HIGH", string(text))
	var p Priority
	require.NoError(t, p.UnmarshalText([]byte("urgent")))
	assert.Equal(t, PriorityHigh, p)

	// SQL: Value stores the underlying string; Scan accepts nil/string/[]byte.
	v, err := PriorityLow.Value()
	require.NoError(t, err)
	assert.Equal(t, "low", v)
	_, err = Priority("nope").Value()
	assert.Error(t, err)

	require.NoError(t, p.Scan("URGENT"))
	assert.Equal(t, PriorityHigh, p)
	require.NoError(t, p.Scan([]byte("LOW")))
	assert.Equal(t, PriorityLow, p)
	require.NoError(t, p.Scan(nil))
	assert.Equal(t, Priority(""), p)
	assert.Error(t, p.Scan(42)) // unsupported source type
}

// TestGeneratedFlagsFullAPI exercises the flags-generated methods left
// uncovered by TestGeneratedFlags: metadata accessors, MustParse, text/SQL
// marshaling, FromValue and undeclared-bit handling.
func TestGeneratedFlagsFullAPI(t *testing.T) {
	// Single declared flags carry no description/alias here, but the
	// accessors must still resolve without panicking.
	assert.Equal(t, "", PermissionRead.Description())
	assert.Empty(t, PermissionRead.Aliases())
	assert.False(t, PermissionRead.HasAlias("anything"))

	assert.Equal(t, PermissionRead.Add(PermissionWrite), MustParsePermission("READ|WRITE"))
	assert.Panics(t, func() { MustParsePermission("READ|NOPE") })

	// Undeclared bits render alongside known ones and fail validation.
	assert.Equal(t, "Permission(8)", Permission(8).String())
	assert.Equal(t, "READ|Permission(8)", Permission(1|8).String())
	_, ok := PermissionFromValue(8)
	assert.False(t, ok)
	rw, ok := PermissionFromValue(1 | 2)
	assert.True(t, ok)
	assert.Equal(t, PermissionRead.Add(PermissionWrite), rw)

	// Text marshaling round-trips the "|"-joined form.
	text, err := PermissionRead.Add(PermissionWrite).MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "READ|WRITE", string(text))
	var p Permission
	require.NoError(t, p.UnmarshalText([]byte("read|delete")))
	assert.Equal(t, PermissionRead.Add(PermissionDelete), p)

	// SQL: Value stores the bit pattern; Scan accepts nil/int64/string/[]byte.
	v, err := PermissionRead.Add(PermissionWrite).Value()
	require.NoError(t, err)
	assert.Equal(t, int64(3), v)
	_, err = Permission(8).Value()
	assert.Error(t, err)

	require.NoError(t, p.Scan(int64(3)))
	assert.Equal(t, PermissionRead.Add(PermissionWrite), p)
	require.NoError(t, p.Scan("WRITE"))
	assert.Equal(t, PermissionWrite, p)
	require.NoError(t, p.Scan([]byte("READ")))
	assert.Equal(t, PermissionRead, p)
	require.NoError(t, p.Scan(nil))
	assert.Equal(t, Permission(0), p)
	assert.Error(t, p.Scan(int64(8))) // undeclared bit
	assert.Error(t, p.Scan(3.5))      // unsupported source type
}

// TestGeneratedUndeclaredValueErrors covers the marshaling guards that reject
// undeclared values so bad data never round-trips silently.
func TestGeneratedUndeclaredValueErrors(t *testing.T) {
	_, err := Status(9).Value()
	assert.Error(t, err)
	_, err = Status(9).MarshalText()
	assert.Error(t, err)
	_, err = Permission(8).MarshalJSON()
	assert.Error(t, err)
	_, err = Permission(8).MarshalText()
	assert.Error(t, err)
}
