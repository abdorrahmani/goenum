// Package status is a worked example of the modern, generated goenum API.
//
// Run `go generate ./...` (or `goenum generate`) in this directory to
// regenerate the *_enum_gen.go files.
package status

//go:generate goenum generate

// Status represents the lifecycle state of an application.
type Status int

const (
	//goenum:name=PENDING
	//goenum:description=Waiting to be processed
	//goenum:alias=WAITING
	StatusPending Status = iota
	//goenum:name=ACTIVE
	//goenum:description=Currently active
	//goenum:alias=RUNNING
	StatusActive
	//goenum:name=DELETED
	//goenum:description=The item has been deleted
	//goenum:alias=REMOVED
	StatusDeleted
)

// Priority is a string-backed enum; values are the wire representation.
type Priority string

const (
	//goenum:name=LOW
	//goenum:description=Low priority task
	PriorityLow Priority = "low"
	//goenum:name=MEDIUM
	//goenum:description=Medium priority task
	PriorityMedium Priority = "medium"
	//goenum:name=HIGH
	//goenum:description=High priority task
	//goenum:alias=URGENT
	PriorityHigh Priority = "high"
)

// Permission is a bit-flag enum.
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
