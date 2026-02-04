package container

import (
	"errors"
	"testing"
	"time"

	"condenser/internal/store/csm"
	"condenser/internal/store/ipam"

	"al.essio.dev/pkg/shellescape"
)

func TestParseImageRef(t *testing.T) {
	svc := &ContainerService{}

	cases := []struct {
		name    string
		input   string
		repo    string
		ref     string
		wantErr bool
	}{
		{name: "plain", input: "ubuntu", repo: "library/ubuntu", ref: "latest"},
		{name: "tag", input: "ubuntu:24.04", repo: "library/ubuntu", ref: "24.04"},
		{name: "explicit repo", input: "library/ubuntu:24.04", repo: "library/ubuntu", ref: "24.04"},
		{name: "digest", input: "nginx@sha256:deadbeef", repo: "library/nginx", ref: "sha256:deadbeef"},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo, ref, err := svc.parseImageRef(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo != tc.repo || ref != tc.ref {
				t.Fatalf("expected %q:%q, got %q:%q", tc.repo, tc.ref, repo, ref)
			}
		})
	}
}

func TestBuildCommand(t *testing.T) {
	svc := &ContainerService{}

	got := svc.buildCommand([]string{"/bin/sh", "-c"}, []string{"echo", "hello"})
	want := shellescape.Quote("/bin/sh") + " " + shellescape.Quote("-c") + " " + shellescape.Quote("echo") + " " + shellescape.Quote("hello")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	got = svc.buildCommand([]string{"/bin/sh"}, []string{"echo", "hello world"})
	want = shellescape.Quote("/bin/sh") + " " + shellescape.Quote("echo") + " " + shellescape.Quote("hello world")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

type fakeCsmHandler struct {
	list []csm.ContainerInfo
	byId map[string]csm.ContainerInfo
	err  error
}

func (f *fakeCsmHandler) StoreContainer(containerId string, state string, pid int, tty bool, repo, ref string, command []string, name string) error {
	return nil
}
func (f *fakeCsmHandler) RemoveContainer(containerId string) error { return nil }
func (f *fakeCsmHandler) UpdateContainer(containerId string, state string, pid int) error {
	return nil
}
func (f *fakeCsmHandler) UpdateSpiffe(containerId string, spiffe string) error { return nil }
func (f *fakeCsmHandler) GetContainerList() ([]csm.ContainerInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}
func (f *fakeCsmHandler) GetContainerById(containerId string) (csm.ContainerInfo, error) {
	if f.err != nil {
		return csm.ContainerInfo{}, f.err
	}
	if info, ok := f.byId[containerId]; ok {
		return info, nil
	}
	return csm.ContainerInfo{}, errors.New("not found")
}
func (f *fakeCsmHandler) IsNameAlreadyUsed(name string) bool { return false }
func (f *fakeCsmHandler) GetContainerIdByName(name string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeCsmHandler) GetContainerNameById(containerId string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeCsmHandler) GetContainerIdAndName(str string) (id, name string, err error) {
	return "", "", errors.New("not implemented")
}
func (f *fakeCsmHandler) GetSpiffeById(containerId string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeCsmHandler) ResolveContainerId(str string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeCsmHandler) IsContainerExist(str string) bool { return false }

type fakeIpamHandler struct {
	pools          []ipam.Pool
	addrById       map[string]string
	allocById      map[string]ipam.Allocation
	errPools       error
	errNetworkInfo error
}

func (f *fakeIpamHandler) Allocate(containerId string, bridge string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeIpamHandler) Release(containerId string) error { return errors.New("not implemented") }
func (f *fakeIpamHandler) GetNetworkList() ([]ipam.NetworkList, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeIpamHandler) GetRuntimeSubnet() (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeIpamHandler) GetDefaultInterface() (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeIpamHandler) GetDefaultInterfaceAddr() (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeIpamHandler) GetBridgeAddr(bridgeInterface string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeIpamHandler) GetDnsProxyInfo() (string, string, []string, error) {
	return "", "", nil, errors.New("not implemented")
}
func (f *fakeIpamHandler) GetContainerAddress(containerId string) (string, string, string, error) {
	return "", "", "", errors.New("not implemented")
}
func (f *fakeIpamHandler) GetInfoByIp(ip string) (string, string, error) {
	return "", "", errors.New("not implemented")
}
func (f *fakeIpamHandler) SetForwardInfo(containerId string, sport, dport int, protocol string) error {
	return errors.New("not implemented")
}
func (f *fakeIpamHandler) GetForwardInfo(containerId string) ([]ipam.ForwardInfo, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeIpamHandler) GetPoolList() ([]ipam.Pool, error) {
	if f.errPools != nil {
		return nil, f.errPools
	}
	return f.pools, nil
}
func (f *fakeIpamHandler) GetNetworkInfoById(containerId string) (string, ipam.Allocation, error) {
	if f.errNetworkInfo != nil {
		return "", ipam.Allocation{}, f.errNetworkInfo
	}
	addr := f.addrById[containerId]
	alloc, ok := f.allocById[containerId]
	if !ok {
		return addr, ipam.Allocation{}, errors.New("not found")
	}
	return addr, alloc, nil
}
func (f *fakeIpamHandler) GetVethById(containerId string) (string, error) {
	return "", errors.New("not implemented")
}

func TestGetContainerList(t *testing.T) {
	created := time.Now()
	csmHandler := &fakeCsmHandler{
		list: []csm.ContainerInfo{{
			ContainerId:   "cid-1",
			ContainerName: "web",
			State:         "running",
			Pid:           123,
			Repository:    "library/nginx",
			Reference:     "latest",
			Command:       []string{"nginx"},
			CreatedAt:     created,
		}},
	}
	ipamHandler := &fakeIpamHandler{
		pools: []ipam.Pool{{
			Interface: "br0",
			Allocations: map[string]ipam.Allocation{
				"10.0.0.10": {
					ContainerId: "cid-1",
					Forwards:    []ipam.ForwardInfo{{HostPort: 80, ContainerPort: 8080, Protocol: "tcp"}},
				},
			},
		}},
	}

	svc := &ContainerService{csmHandler: csmHandler, ipamHandler: ipamHandler}
	list, err := svc.GetContainerList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 container, got %d", len(list))
	}
	if list[0].Address != "10.0.0.10" {
		t.Fatalf("expected address, got %q", list[0].Address)
	}
	if len(list[0].Forwards) != 1 || list[0].Forwards[0].HostPort != 80 {
		t.Fatalf("unexpected forwards: %+v", list[0].Forwards)
	}
}

func TestGetContainerList_NoAllocation(t *testing.T) {
	csmHandler := &fakeCsmHandler{list: []csm.ContainerInfo{{ContainerId: "cid-1", ContainerName: "web"}}}
	ipamHandler := &fakeIpamHandler{pools: []ipam.Pool{{Interface: "br0", Allocations: map[string]ipam.Allocation{}}}}

	svc := &ContainerService{csmHandler: csmHandler, ipamHandler: ipamHandler}
	list, err := svc.GetContainerList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 container, got %d", len(list))
	}
	if list[0].Address != "" {
		t.Fatalf("expected empty address, got %q", list[0].Address)
	}
	if list[0].Forwards != nil {
		t.Fatalf("expected nil forwards, got %+v", list[0].Forwards)
	}
}

func TestGetContainerById(t *testing.T) {
	created := time.Now()
	csmHandler := &fakeCsmHandler{
		byId: map[string]csm.ContainerInfo{
			"cid-1": {
				ContainerId: "cid-1",
				State:       "running",
				Pid:         42,
				Repository:  "library/nginx",
				Reference:   "latest",
				Command:     []string{"nginx"},
				CreatedAt:   created,
			},
		},
	}
	ipamHandler := &fakeIpamHandler{
		addrById: map[string]string{"cid-1": "10.0.0.20"},
		allocById: map[string]ipam.Allocation{
			"cid-1": {Forwards: []ipam.ForwardInfo{{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"}}},
		},
	}

	svc := &ContainerService{csmHandler: csmHandler, ipamHandler: ipamHandler}
	state, err := svc.GetContainerById("cid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Address != "10.0.0.20" {
		t.Fatalf("expected address, got %q", state.Address)
	}
	if len(state.Forwards) != 1 || state.Forwards[0].ContainerPort != 80 {
		t.Fatalf("unexpected forwards: %+v", state.Forwards)
	}
}

func TestGetContainerById_Errors(t *testing.T) {
	svc := &ContainerService{
		csmHandler:  &fakeCsmHandler{err: errors.New("csm fail")},
		ipamHandler: &fakeIpamHandler{},
	}
	if _, err := svc.GetContainerById("cid-1"); err == nil {
		t.Fatalf("expected error from csm handler")
	}

	svc = &ContainerService{
		csmHandler:  &fakeCsmHandler{byId: map[string]csm.ContainerInfo{"cid-1": {ContainerId: "cid-1"}}},
		ipamHandler: &fakeIpamHandler{errNetworkInfo: errors.New("ipam fail")},
	}
	if _, err := svc.GetContainerById("cid-1"); err == nil {
		t.Fatalf("expected error from ipam handler")
	}
}

func TestGetContainerState(t *testing.T) {
	svc := &ContainerService{
		csmHandler: &fakeCsmHandler{byId: map[string]csm.ContainerInfo{"cid-1": {ContainerId: "cid-1", State: "running"}}},
	}
	state, err := svc.getContainerState("cid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "running" {
		t.Fatalf("expected running, got %q", state)
	}
}
