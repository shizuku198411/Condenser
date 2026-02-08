package psm

import "time"

type PodInfo struct {
	PodId       string            `json:"podId"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	UID         string            `json:"uid"`
	State       string            `json:"state"`
	OwnerPid    int               `json:"ownerPid"`
	NetworkNS   string            `json:"networkNS"`
	IPCNS       string            `json:"ipcNS"`
	UTSNS       string            `json:"utsNS"`
	UserNS      string            `json:"userNS"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	StartedAt   time.Time         `json:"startedAt"`
	StoppedAt   time.Time         `json:"stoppedAt"`
}

type PodState struct {
	Version string             `json:"version"`
	Pods    map[string]PodInfo `json:"pods"`
}
