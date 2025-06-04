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

	// JSON operations with different formats
	// Default format (name only)
	jsonData, _ := json.Marshal(StatusActive)
	fmt.Printf("Default JSON: %s\n", jsonData)

	// Value format
	StatusActive.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
	jsonData, _ = json.Marshal(StatusActive)
	fmt.Printf("Value JSON: %s\n", jsonData)

	// Full format
	StatusActive.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
	jsonData, _ = json.Marshal(StatusActive)
	fmt.Printf("Full JSON: %s\n", jsonData)

	// Unmarshal examples
	var status Status
	status.EnumBase = &EnumBase{}

	// Unmarshal name format
	_ = json.Unmarshal([]byte(`"PENDING"`), &status)
	fmt.Printf("Unmarshaled name: %s\n", status.String())

	// Unmarshal value format
	status.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatValue})
	_ = json.Unmarshal([]byte(`1`), &status)
	fmt.Printf("Unmarshaled value: %v\n", status.Value())

	// Unmarshal full format
	status.SetJSONConfig(&EnumJSONConfig{Format: JSONFormatFull})
	fullJSON := `{"name":"ACTIVE","value":1,"description":"The item is currently active","aliases":["RUNNING","LIVE"]}`
	_ = json.Unmarshal([]byte(fullJSON), &status)
	fmt.Printf("Unmarshaled full: %s (value: %v, desc: %s)\n",
		status.String(), status.Value(), status.Description())

	// New utility methods examples
	fmt.Printf("All status names: %v\n", StatusEnumSet.Names())
	fmt.Printf("Status map: %v\n", StatusEnumSet.Map())

	// Filter active and pending statuses
	activeStatuses := StatusEnumSet.Filter(func(s Status) bool {
		return s.Value().(int) < 2 // Filter statuses with value less than 2
	})
	fmt.Printf("Active statuses: %v\n", activeStatuses)

	// Composite enum examples
	var (
		PermissionRead  = NewCompositeEnumBase(0, "READ", "Read permission")
		PermissionWrite = NewCompositeEnumBase(1, "WRITE", "Write permission")
		PermissionExec  = NewCompositeEnumBase(2, "EXEC", "Execute permission")
	)

	// Combine permissions
	allPermissions := PermissionRead.Or(PermissionWrite).Or(PermissionExec)
	fmt.Printf("All permissions: %s (value: %v)\n", allPermissions.String(), allPermissions.Value())

	// Check multiple flags
	fmt.Printf("Has read and write: %v\n", allPermissions.HasAllFlags(PermissionRead, PermissionWrite))
	fmt.Printf("Has read and exec: %v\n", allPermissions.HasAllFlags(PermissionRead, PermissionExec))

	// Remove permission
	readWriteOnly := allPermissions.RemoveFlag(PermissionExec)
	fmt.Printf("Read and write only: %s (value: %v)\n", readWriteOnly.String(), readWriteOnly.Value())
	fmt.Printf("Still has exec: %v\n", readWriteOnly.HasFlag(PermissionExec))
}
