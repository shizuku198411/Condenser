package ipam

import (
	"condenser/internal/utils"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

func NewIpamStore(path string) *IpamStore {
	return &IpamStore{
		path:              path,
		filesystemHandler: utils.NewFilesystemExecutor(),
	}
}

type IpamStore struct {
	path              string
	mu                sync.Mutex
	filesystemHandler utils.FilesystemHandler
}

func (s *IpamStore) withLock(fn func(st *IpamState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockPath := s.path + ".lock"
	if err := s.filesystemHandler.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	lf, err := s.filesystemHandler.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lf.Close()

	if err := s.filesystemHandler.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer s.filesystemHandler.Flock(int(lf.Fd()), syscall.LOCK_UN)

	st, err := s.loadOrInit()
	if err != nil {
		return err
	}

	if err := fn(st); err != nil {
		return err
	}

	return s.atomicSave(st)
}

func (s *IpamStore) loadOrInit() (*IpamState, error) {
	b, err := s.filesystemHandler.ReadFile(s.path)
	if err != nil {
		if s.filesystemHandler.IsNotExist(err) {
			// ipam state file not exist
			return &IpamState{
				Version:     "0.1.0",
				Allocations: map[string]Allocation{},
			}, nil
		}
		return nil, err
	}

	var st IpamState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("ipam state json broken: %w", err)
	}
	if st.Allocations == nil {
		st.Allocations = map[string]Allocation{}
	}
	return &st, nil
}

func (s *IpamStore) atomicSave(st *IpamState) error {
	tmp := s.path + ".tmp"

	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	f, err := s.filesystemHandler.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return s.filesystemHandler.Rename(tmp, s.path)
}

func (s *IpamStore) SetConfig(subnetCIDR string, gateway string) error {
	// validate address format
	_, _, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}
	if net.ParseIP(gateway) == nil {
		return fmt.Errorf("invalid gateway address")
	}

	return s.withLock(func(st *IpamState) error {
		if st.Subnet != "" && st.Subnet != subnetCIDR {
			return fmt.Errorf("ipam subnet already set: %s", st.Subnet)
		}
		st.Version = "0.1.0"
		st.Subnet = subnetCIDR
		st.Gateway = gateway
		if st.Allocations == nil {
			st.Allocations = map[string]Allocation{}
		}
		return nil
	})
}
