package bottle

import (
	"condenser/internal/core/container"
	"condenser/internal/core/policy"
	"condenser/internal/store/bsm"
	"condenser/internal/store/csm"
	"condenser/internal/store/ipam"
	"condenser/internal/utils"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func NewBottleService() *BottleService {
	return &BottleService{
		containerService: container.NewContaierService(),
		bsmHandler:       bsm.NewBsmManager(bsm.NewBsmStore(utils.BsmStorePath)),
		csmHandler:       csm.NewCsmManager(csm.NewCsmStore(utils.CsmStorePath)),
		ipamHandler:      ipam.NewIpamManager(ipam.NewIpamStore(utils.IpamStorePath)),
		policyHandler:    policy.NewwServicePolicy(),
	}
}

type BottleService struct {
	containerService container.ContainerServiceHandler
	bsmHandler       bsm.BsmHandler
	csmHandler       csm.CsmHandler
	ipamHandler      ipam.IpamHandler
	policyHandler    policy.PolicyServiceHandler
}

func (s *BottleService) DecodeSpec(yamlBytes []byte) (*BottleSpec, error) {
	var spec BottleSpec
	if err := yaml.Unmarshal(yamlBytes, &spec); err != nil {
		return nil, err
	}
	if spec.Bottle.Name == "" {
		return nil, fmt.Errorf("bottle.name is required")
	}
	if len(spec.Services) == 0 {
		return nil, fmt.Errorf("services is required")
	}
	return &spec, nil
}

func (s *BottleService) BuildStartOrder(spec *BottleSpec) ([]string, error) {
	inDegree := make(map[string]int, len(spec.Services))
	graph := make(map[string][]string, len(spec.Services))

	for name := range spec.Services {
		inDegree[name] = 0
	}

	for svcName, svc := range spec.Services {
		for _, dep := range svc.DependsOn {
			if _, ok := spec.Services[dep]; !ok {
				return nil, fmt.Errorf("service %q depends on unknown service %q", svcName, dep)
			}
			graph[dep] = append(graph[dep], svcName)
			inDegree[svcName]++
		}
	}

	order := make([]string, 0, len(spec.Services))
	queue := make([]string, 0, len(spec.Services))
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		order = append(order, n)
		for _, next := range graph[n] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(spec.Services) {
		return nil, fmt.Errorf("dependency cycle detected")
	}
	return order, nil
}

func (s *BottleService) Create(bottleIdOrName string) (string, error) {
	bottleId, err := s.bsmHandler.ResolveBottleId(bottleIdOrName)
	if err != nil {
		return "", err
	}

	info, err := s.bsmHandler.GetBottleById(bottleId)
	if err != nil {
		return "", err
	}
	if len(info.StartOrder) == 0 {
		return "", fmt.Errorf("start order is empty")
	}

	containers := make(map[string]string, len(info.Containers))
	for k, v := range info.Containers {
		containers[k] = v
	}

	var created []string
	for _, serviceName := range info.StartOrder {
		containerId := ""
		if containers != nil {
			containerId = containers[serviceName]
		}
		if containerId != "" {
			continue
		}

		spec, ok := info.Services[serviceName]
		if !ok {
			return "", fmt.Errorf("service spec not found: %s", serviceName)
		}
		env, err := s.resolveEnvWithDeps(info, containers, serviceName, spec.Env)
		if err != nil {
			s.rollbackCreatedContainers(created, containers, bottleId)
			return "", err
		}
		createParam := container.ServiceCreateModel{
			Image:    spec.Image,
			Command:  spec.Command,
			Port:     spec.Ports,
			Mount:    spec.Mount,
			Env:      env,
			Network:  spec.Network,
			Tty:      spec.Tty,
			Name:     buildContainerName(info.BottleName, serviceName),
			BottleId: bottleId,
		}
		containerId, err = s.containerService.Create(createParam)
		if err != nil {
			s.rollbackCreatedContainers(created, containers, bottleId)
			return "", err
		}

		containers[serviceName] = containerId
		created = append(created, containerId)
		if err := s.bsmHandler.UpdateBottleContainer(bottleId, serviceName, containerId); err != nil {
			s.rollbackCreatedContainers(created, containers, bottleId)
			return "", err
		}
	}

	return bottleId, nil
}

func (s *BottleService) Start(bottleIdOrName string) (string, error) {
	bottleId, err := s.bsmHandler.ResolveBottleId(bottleIdOrName)
	if err != nil {
		return "", err
	}

	info, err := s.bsmHandler.GetBottleById(bottleId)
	if err != nil {
		return "", err
	}
	if len(info.StartOrder) == 0 {
		return "", fmt.Errorf("start order is empty")
	}

	for _, serviceName := range info.StartOrder {
		containerId := ""
		if info.Containers != nil {
			containerId = info.Containers[serviceName]
		}
		if containerId == "" {
			return "", fmt.Errorf("container for service %s not created", serviceName)
		}

		spec, ok := info.Services[serviceName]
		if !ok {
			return "", fmt.Errorf("service spec not found: %s", serviceName)
		}
		if _, err := s.containerService.Start(container.ServiceStartModel{
			ContainerId: containerId,
			Tty:         spec.Tty,
			OpBottle:    true,
		}); err != nil {
			return "", err
		}
	}

	return bottleId, nil
}

func (s *BottleService) Stop(bottleIdOrName string) (string, error) {
	bottleId, err := s.bsmHandler.ResolveBottleId(bottleIdOrName)
	if err != nil {
		return "", err
	}

	info, err := s.bsmHandler.GetBottleById(bottleId)
	if err != nil {
		return "", err
	}
	if len(info.StartOrder) == 0 {
		return "", fmt.Errorf("start order is empty")
	}

	for i := len(info.StartOrder) - 1; i >= 0; i-- {
		serviceName := info.StartOrder[i]
		containerId := ""
		if info.Containers != nil {
			containerId = info.Containers[serviceName]
		}
		if containerId == "" {
			return "", fmt.Errorf("container for service %s not found", serviceName)
		}
		state, err := s.getContainerState(containerId)
		if err != nil {
			return "", err
		}
		if state != "running" {
			continue
		}
		if _, err := s.containerService.Stop(container.ServiceStopModel{
			ContainerId: containerId,
			OpBottle:    true,
		}); err != nil {
			return "", err
		}
	}

	return bottleId, nil
}

func (s *BottleService) Delete(bottleIdOrName string) (string, error) {
	bottleId, err := s.bsmHandler.ResolveBottleId(bottleIdOrName)
	if err != nil {
		return "", err
	}

	info, err := s.bsmHandler.GetBottleById(bottleId)
	if err != nil {
		return "", err
	}
	if len(info.StartOrder) == 0 {
		return "", fmt.Errorf("start order is empty")
	}

	for i := len(info.StartOrder) - 1; i >= 0; i-- {
		serviceName := info.StartOrder[i]
		containerId := ""
		if info.Containers != nil {
			containerId = info.Containers[serviceName]
		}
		if containerId == "" {
			return "", fmt.Errorf("container for service %s not found", serviceName)
		}

		state, err := s.getContainerState(containerId)
		if err != nil {
			return "", err
		}
		if state == "running" {
			if _, err := s.containerService.Stop(container.ServiceStopModel{
				ContainerId: containerId,
				OpBottle:    true,
			}); err != nil {
				return "", err
			}
		}

		if _, err := s.containerService.Delete(container.ServiceDeleteModel{
			ContainerId: containerId,
			OpBottle:    true,
		}); err != nil {
			return "", err
		}
	}

	if err := s.removeBottlePolicies(info.Policies); err != nil {
		return "", err
	}

	if err := s.bsmHandler.RemoveBottle(bottleId); err != nil {
		return "", err
	}
	return bottleId, nil
}

func (s *BottleService) getContainerState(containerId string) (string, error) {
	info, err := s.csmHandler.GetContainerById(containerId)
	if err != nil {
		return "", err
	}
	return info.State, nil
}

func buildContainerName(bottleName string, serviceName string) string {
	return bottleName + "-" + serviceName
}

func (s *BottleService) removeBottlePolicies(policies []bsm.PolicyInfo) error {
	for _, p := range policies {
		if p.Id == "" {
			continue
		}
		if err := s.policyHandler.RemoveUserPolicy(policy.ServiceRemovePolicyModel{Id: p.Id}); err != nil {
			return err
		}
		// commit
		if err := s.policyHandler.CommitPolicy(); err != nil {
			return err
		}
	}
	return nil
}

func (s *BottleService) rollbackCreatedContainers(created []string, containers map[string]string, bottleId string) {
	for i := len(created) - 1; i >= 0; i-- {
		_, _ = s.containerService.Delete(container.ServiceDeleteModel{
			ContainerId: created[i],
		})
	}
	for _, id := range created {
		for name, cid := range containers {
			if cid == id {
				delete(containers, name)
			}
		}
	}
	_ = s.bsmHandler.UpdateBottleContainers(bottleId, containers)
}

func (s *BottleService) resolveEnvWithDeps(info bsm.BottleInfo, containers map[string]string, serviceName string, env []string) ([]string, error) {
	spec, ok := info.Services[serviceName]
	if !ok || len(spec.DependsOn) == 0 || len(env) == 0 {
		return env, nil
	}

	depAddrs := map[string]string{}
	for _, dep := range spec.DependsOn {
		containerId := ""
		if containers != nil {
			containerId = containers[dep]
		}
		if containerId == "" {
			name := buildContainerName(info.BottleName, dep)
			id, err := s.csmHandler.GetContainerIdByName(name)
			if err != nil {
				return nil, fmt.Errorf("container for dependency %s not found", dep)
			}
			containerId = id
		}
		addr, _, err := s.ipamHandler.GetNetworkInfoById(containerId)
		if err != nil {
			return nil, err
		}
		depAddrs[dep] = addr
	}

	out := make([]string, 0, len(env))
	for _, kv := range env {
		key, val, ok := strings.Cut(kv, "=")
		if !ok {
			out = append(out, kv)
			continue
		}
		for dep, addr := range depAddrs {
			if strings.HasPrefix(val, dep+":") {
				val = addr + val[len(dep):]
				break
			}
			if val == dep {
				val = addr
				break
			}
		}
		out = append(out, key+"="+val)
	}
	return out, nil
}
