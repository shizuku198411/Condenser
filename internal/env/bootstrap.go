package env

import (
	"bufio"
	"condenser/internal/csm"
	"condenser/internal/ipam"
	"condenser/internal/utils"
	"fmt"
	"os"
	"strings"
)

func NewBootstrapManager() *BootstrapManager {
	return &BootstrapManager{
		filesystemHandler: utils.NewFilesystemExecutor(),
		commandFactory:    utils.NewCommandFactory(),
		ipamStoreHandler:  ipam.NewIpamStore(IpamStorePath),
		csmStoreHandler:   csm.NewCsmStore(CsmStorePath),
	}
}

type BootstrapManager struct {
	filesystemHandler utils.FilesystemHandler
	commandFactory    utils.CommandFactory
	ipamStoreHandler  ipam.IpamStoreHandler
	csmStoreHandler   csm.CsmStoreHandler
}

func (m *BootstrapManager) SetupRuntime() error {
	// 1. create runtime directory
	if err := m.setupRuntimeDirectory(); err != nil {
		return err
	}

	// 2. setup cgroup
	if err := m.setupCgroup(); err != nil {
		return err
	}

	// 3. setup network
	if err := m.setupNetwork(); err != nil {
		return err
	}

	// 4. setup IPAM (IP Address Managemant)
	if err := m.setupIpam(); err != nil {
		return err
	}

	// 5. setup CSM (Container State Management)
	if err := m.setupCsm(); err != nil {
		return err
	}

	return nil
}

func (m *BootstrapManager) setupRuntimeDirectory() error {
	dirs := []string{
		ContainerRootDir,
		ImageRootDir,
		LayerRootDir,
		StoreDir,
	}
	for _, dir := range dirs {
		if err := m.filesystemHandler.MkdirAll(dir, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (m *BootstrapManager) setupCgroup() error {
	// 1. create cgroup runtime directory
	if err := m.setupCgroupDirectory(); err != nil {
		return err
	}

	// 2. enable controllers
	if err := m.enableCgroupControllers(); err != nil {
		return err
	}

	return nil
}

func (m *BootstrapManager) setupCgroupDirectory() error {
	dir := CgroupRuntimeDir
	if err := m.filesystemHandler.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return nil
}

func (m *BootstrapManager) enableCgroupControllers() error {
	// get current enabled control
	enabled, err := m.readCgroupEnabledControllers()
	if err != nil {
		return nil
	}

	controllers := []string{
		"cpu",
		"memory",
	}
	for _, c := range controllers {
		if enabled[c] {
			continue
		}
		if err := m.writeCgroupController("+" + c); err != nil {
			return err
		}
	}

	return nil
}

func (m *BootstrapManager) readCgroupEnabledControllers() (map[string]bool, error) {
	subtreePath := CgroupSubtreeControlPath
	f, err := m.filesystemHandler.Open(subtreePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	enabled := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		for _, name := range fields {
			enabled[name] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return enabled, nil
}

func (m *BootstrapManager) writeCgroupController(token string) error {
	subtreePath := CgroupSubtreeControlPath
	f, err := m.filesystemHandler.OpenFile(subtreePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s\n", token); err != nil {
		return err
	}
	return nil
}

func (m *BootstrapManager) setupNetwork() error {
	// 1. create bridge interface
	if err := m.createBridgeInterface(); err != nil {
		return err
	}

	// 2. setup masquerade
	if err := m.createMasqueradeRule(); err != nil {
		return err
	}

	return nil
}

func (m *BootstrapManager) createBridgeInterface() error {
	// check if bridge interface already exist
	check := m.commandFactory.Command("ip", "link", "show", BridgeInterface)
	if err := check.Run(); err == nil {
		// bridge interface already exist
		return nil
	}

	// create bridge interface
	add := m.commandFactory.Command("ip", "link", "add", BridgeInterface, "type", "bridge")
	if err := add.Run(); err != nil {
		return fmt.Errorf("ip link add: %w", err)
	}

	// assign address
	assign := m.commandFactory.Command("ip", "addr", "add", BridgeInterfaceAddr, "dev", BridgeInterface)
	if err := assign.Run(); err != nil {
		return fmt.Errorf("ip addr add: %w", err)
	}

	// up link
	up := m.commandFactory.Command("ip", "link", "set", BridgeInterface, "up")
	if err := up.Run(); err != nil {
		return fmt.Errorf("ip link up: %w", err)
	}
	return nil
}

func (m *BootstrapManager) createMasqueradeRule() error {
	hostInterface, err := m.getDefaultInterfaceIpv4()
	if err != nil {
		return err
	}
	// check if rule already exist
	check := m.commandFactory.Command("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", RuntimeSubnet, "-o", hostInterface, "-j", "MASQUERADE")
	if err := check.Run(); err == nil {
		// rule already exist
		return nil
	}

	// add rule
	add := m.commandFactory.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", RuntimeSubnet, "-o", hostInterface, "-j", "MASQUERADE")
	if err := add.Run(); err != nil {
		return fmt.Errorf("iptables add: %w", err)
	}
	return nil
}

func (m *BootstrapManager) getDefaultInterfaceIpv4() (string, error) {
	cmd := m.commandFactory.Command("ip", "-4", "route", "show", "default")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run ip route: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", fmt.Errorf("no defauilt route found (ipv4)")
	}

	// retrieve device name
	fields := strings.Fields(lines[0])
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "dev" {
			return fields[i+1], nil
		}
	}
	return "", fmt.Errorf("cannot find dev in: %q", lines[0])
}

func (m *BootstrapManager) setupIpam() error {
	subnet := RuntimeSubnet
	gateway := strings.Split(BridgeInterfaceAddr, "/")[0]

	return m.ipamStoreHandler.SetConfig(subnet, gateway)
}

func (m *BootstrapManager) setupCsm() error {
	return m.csmStoreHandler.SetContainerState()
}
