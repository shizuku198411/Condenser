package container

type ServiceCreateModel struct {
	Image   string
	Command []string
}

type ServiceStartModel struct {
	ContainerId string
	Interactive bool
}

type ServiceDeleteModel struct {
	ContainerId string
}

type ServiceStopModel struct {
	ContainerId string
}
