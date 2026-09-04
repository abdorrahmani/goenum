package goenum_test

import (
	"encoding/json"
	"testing"

	goenum "github.com/abdorrahmani/goenum"
	"github.com/abdorrahmani/goenum/example/status"
)

// Legacy architecture: EnumBase + EnumSet, as consumers used before the
// generated/generic redesign.

type LegacyStatus struct {
	*goenum.EnumBase
}

var (
	LegacyPending = LegacyStatus{goenum.NewEnumBase(0, "PENDING", "Waiting to be processed", "WAITING")}
	LegacyActive  = LegacyStatus{goenum.NewEnumBase(1, "ACTIVE", "Currently active", "RUNNING")}
	LegacyDeleted = LegacyStatus{goenum.NewEnumBase(2, "DELETED", "The item has been deleted", "REMOVED")}
)

var LegacyStatuses = goenum.NewEnumSet[LegacyStatus]().
	Register(LegacyPending).
	Register(LegacyActive).
	Register(LegacyDeleted)

func (s LegacyStatus) MarshalJSON() ([]byte, error) { return s.EnumBase.MarshalJSON() }

func (s *LegacyStatus) UnmarshalJSON(data []byte) error {
	if s.EnumBase == nil {
		s.EnumBase = &goenum.EnumBase{}
	}
	return s.EnumBase.UnmarshalJSON(data)
}

// Parsing

func BenchmarkParseByNameLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = LegacyStatuses.Parse("ACTIVE")
	}
}

func BenchmarkParseByNameGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = status.ParseStatus("ACTIVE")
	}
}

func BenchmarkParseByNameGeneric(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = goenum.Parse[status.Status]("ACTIVE")
	}
}

func BenchmarkParseAliasLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = LegacyStatuses.Parse("RUNNING")
	}
}

func BenchmarkParseAliasGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = status.ParseStatus("RUNNING")
	}
}

// Lookup by value

func BenchmarkLookupByValueLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = LegacyStatuses.GetByValue(1)
	}
}

func BenchmarkLookupByValueGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = status.StatusFromValue(1)
	}
}

// String / IsValid

func BenchmarkStringLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = LegacyActive.String()
	}
}

func BenchmarkStringGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = status.StatusActive.String()
	}
}

func BenchmarkIsValidLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = LegacyActive.IsValid()
	}
}

func BenchmarkIsValidGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = status.StatusActive.IsValid()
	}
}

// JSON

func BenchmarkJSONMarshalLegacy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = json.Marshal(LegacyActive)
	}
}

func BenchmarkJSONMarshalGenerated(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = json.Marshal(status.StatusActive)
	}
}

func BenchmarkJSONUnmarshalLegacy(b *testing.B) {
	data := []byte(`"ACTIVE"`)
	b.ReportAllocs()
	for b.Loop() {
		var s LegacyStatus
		_ = json.Unmarshal(data, &s)
	}
}

func BenchmarkJSONUnmarshalGenerated(b *testing.B) {
	data := []byte(`"ACTIVE"`)
	b.ReportAllocs()
	for b.Loop() {
		var s status.Status
		_ = json.Unmarshal(data, &s)
	}
}
