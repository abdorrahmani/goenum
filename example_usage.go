package goenum

import (
	"encoding/json"
	"fmt"
)

// Status represents an example enum type
type Status struct {
	*EnumBase
}

var (
	StatusPending = Status{NewEnumBase(0, "PENDING", "The item is waiting to be processed", "WAITING")}
	StatusActive  = Status{NewEnumBase(1, "ACTIVE", "The item is currently active", "RUNNING", "LIVE")}
	StatusDeleted = Status{NewEnumBase(2, "DELETED", "The item has been deleted", "REMOVED")}
)

var StatusEnumSet = NewEnumSet[Status]()

func init() {
	// Using chainable Register method
	StatusEnumSet.Register(StatusPending).
		Register(StatusActive).
		Register(StatusDeleted)
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

// Example demonstrates the usage of the improved enum package
func Example() {
	// Basic enum operations
	fmt.Printf("Status: %s, Value: %v, Description: %s\n",
		StatusActive.String(),
		StatusActive.Value(),
		StatusActive.Description())

	// Check aliases
	fmt.Printf("Has alias 'RUNNING': %v\n", StatusActive.HasAlias("RUNNING"))
	fmt.Printf("All aliases: %v\n", StatusActive.Aliases())

	// EnumSet operations
	if status, exists := StatusEnumSet.GetByName("ACTIVE"); exists {
		fmt.Printf("Found by name: %s\n", status.String())
	}

	if status, exists := StatusEnumSet.GetByValue(1); exists {
		fmt.Printf("Found by value: %s\n", status.String())
	}

	// Try finding by alias
	if status, exists := StatusEnumSet.GetByName("WAITING"); exists {
		fmt.Printf("Found by alias: %s\n", status.String())
	}

	// JSON operations
	jsonData, _ := json.Marshal(StatusActive)
	fmt.Printf("JSON: %s\n", jsonData)

	var status Status
	_ = json.Unmarshal([]byte(`"PENDING"`), &status)
	fmt.Printf("Unmarshaled: %s\n", status.String())
}
