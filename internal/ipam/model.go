package ipam

import (
	"time"
)

type Allocation struct {
	ContainerId   string    `json:"containerId"`
	InterfaceName string    `json:"interfaceName,omitempty"`
	AssignedAt    time.Time `json:"assignedAt"`
}

type IpamState struct {
	Version       string                `json:"version"`
	Subnet        string                `json:"subnet"`
	Gateway       string                `json:"gateway"`
	Allocations   map[string]Allocation `json:"allocations"`
	LastAllocated string                `json:"lastAllocated,omitempty"`
}
