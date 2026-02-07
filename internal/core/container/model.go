package container

import "time"

type ServiceCreateModel struct {
	Image    string
	Os       string
	Arch     string
	Command  []string
	Port     []string
	Mount    []string
	Env      []string
	Network  string
	Tty      bool
	Name     string
	BottleId string
}

type ServiceStartModel struct {
	ContainerId string
	Tty         bool
	OpBottle    bool
}

type ServiceDeleteModel struct {
	ContainerId string
	OpBottle    bool
}

type ServiceStopModel struct {
	ContainerId string
	OpBottle    bool
}

type ServiceExecModel struct {
	ContainerId string
	Tty         bool
	Entrypoint  []string
}

type ForwardInfo struct {
	HostPort      int    `json:"source"`
	ContainerPort int    `json:"destination"`
	Protocol      string `json:"protocol"`
}

type ContainerState struct {
	ContainerId string   `json:"containerId"`
	Name        string   `json:"name"`
	State       string   `json:"state"`
	Pid         int      `json:"pid"`
	Repository  string   `json:"imageRepository"`
	Reference   string   `json:"imageReference"`
	Command     []string `json:"command"`

	Address  string        `json:"address"`
	Forwards []ForwardInfo `json:"forwards"`

	CreatingAt time.Time `json:"creatingAt"`
	CreatedAt  time.Time `json:"createdAt"`
	StartedAt  time.Time `json:"statedAt"`
	StoppedAt  time.Time `json:"stoppedAt"`
}
