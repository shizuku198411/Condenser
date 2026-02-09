package pod

import "condenser/internal/core/container"

// == service: remove pod sandbox ==
func (s *PodService) Remove(podId string) (string, error) {
	containers, err := s.containerHandler.GetContainersByPodId(podId)
	if err != nil {
		return "", err
	}
	var infra container.ContainerState
	var hasInfra bool
	for _, c := range containers {
		if s.isPodInfraName(c.Name) {
			infra = c
			hasInfra = true
			continue
		}
		if _, err := s.containerHandler.Delete(container.ServiceDeleteModel{ContainerId: c.ContainerId}); err != nil {
			return "", err
		}
	}
	if hasInfra {
		if _, err := s.containerHandler.Stop(container.ServiceStopModel{ContainerId: infra.ContainerId}); err != nil {
			return "", err
		}
		if _, err := s.containerHandler.Delete(container.ServiceDeleteModel{ContainerId: infra.ContainerId}); err != nil {
			return "", err
		}
	}
	if err := s.psmHandler.RemovePod(podId); err != nil {
		return "", err
	}
	return podId, nil
}
