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
