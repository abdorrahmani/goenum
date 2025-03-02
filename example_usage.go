package goenum

import "encoding/json"

// Status represents an example enum type
type Status struct {
	*EnumBase
}

var (
	StatusPending = Status{&EnumBase{value: 0, name: "PENDING"}}
	StatusActive  = Status{&EnumBase{value: 1, name: "ACTIVE"}}
	StatusDeleted = Status{&EnumBase{value: 2, name: "DELETED"}}
)

var StatusEnumSet = NewEnumSet[Status]()

func init() {
	StatusEnumSet.Register(StatusPending)
	StatusEnumSet.Register(StatusActive)
	StatusEnumSet.Register(StatusDeleted)
}

// MarshalJSON implements JSON marshaling for Status
func (s Status) MarshalJSON() ([]byte, error) {
	if s.EnumBase == nil {
		return json.Marshal("")
	}
	return s.EnumBase.MarshalJSON()
}

// UnmarshalJSON implements JSON unmarshaling for Status
func (s *Status) UnmarshalJSON(data []byte) error {
	if s.EnumBase == nil {
		s.EnumBase = &EnumBase{}
	}
	return s.EnumBase.UnmarshalJSON(data)
}
