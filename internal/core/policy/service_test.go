package policy

import (
	"errors"
	"testing"

	"condenser/internal/store/csm"
)

type fakeCsmHandler struct {
	byName map[string]string
	byId   map[string]string
}

func (f *fakeCsmHandler) StoreContainer(containerId string, state string, pid int, tty bool, repo, ref string, command []string, name string) error {
	return nil
}
func (f *fakeCsmHandler) RemoveContainer(containerId string) error { return nil }
func (f *fakeCsmHandler) UpdateContainer(containerId string, state string, pid int) error {
	return nil
}
func (f *fakeCsmHandler) UpdateSpiffe(containerId string, spiffe string) error { return nil }
func (f *fakeCsmHandler) GetContainerList() ([]csm.ContainerInfo, error)       { return nil, nil }
func (f *fakeCsmHandler) GetContainerById(containerId string) (csm.ContainerInfo, error) {
	return csm.ContainerInfo{}, errors.New("not found")
}
func (f *fakeCsmHandler) IsNameAlreadyUsed(name string) bool { return false }
func (f *fakeCsmHandler) GetContainerIdByName(name string) (string, error) {
	if id, ok := f.byName[name]; ok {
		return id, nil
	}
	return "", errors.New("not found")
}
func (f *fakeCsmHandler) GetContainerNameById(containerId string) (string, error) {
	if name, ok := f.byId[containerId]; ok {
		return name, nil
	}
	return "", errors.New("not found")
}
func (f *fakeCsmHandler) GetContainerIdAndName(str string) (id, name string, err error) {
	return "", "", errors.New("not implemented")
}
func (f *fakeCsmHandler) GetSpiffeById(containerId string) (string, error) {
	return "", errors.New("not found")
}
func (f *fakeCsmHandler) ResolveContainerId(str string) (string, error) {
	return "", errors.New("not found")
}
func (f *fakeCsmHandler) IsContainerExist(str string) bool { return false }

func TestResolveContainerNameAndInfo(t *testing.T) {
	service := &ServicePolicy{
		csmHandler: &fakeCsmHandler{
			byName: map[string]string{"web": "id-1"},
			byId:   map[string]string{"id-2": "db"},
		},
	}

	id, name := service.resolveContainerNameAndInfo("web")
	if id != "id-1" || name != "web" {
		t.Fatalf("expected id-1/web, got %q/%q", id, name)
	}

	id, name = service.resolveContainerNameAndInfo("id-2")
	if id != "id-2" || name != "db" {
		t.Fatalf("expected id-2/db, got %q/%q", id, name)
	}

	id, name = service.resolveContainerNameAndInfo("unknown")
	if id != "" || name != "unknown" {
		t.Fatalf("expected empty id and name passthrough, got %q/%q", id, name)
	}
}
