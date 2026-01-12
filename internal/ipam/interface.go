package ipam

type IpamStoreHandler interface {
	SetConfig(subnetCIDR string, gateway string) error
}

type IpamHandler interface {
	Allocate(containerId string) (string, error)
	Release(containerId string) error
}
