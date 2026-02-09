package pod

import "condenser/internal/core/container"

// == service: start pod sandbox ==
func (s *PodService) Start(podId string) (string, error) {
	containers, err := s.containerHandler.GetContainersByPodId(podId)
	if err != nil {
		return "", err
	}

	for _, c := range containers {
		if s.isPodInfraName(c.Name) {
			if c.State == "running" {
				continue
			}
			if _, err := s.containerHandler.Start(container.ServiceStartModel{ContainerId: c.ContainerId, Tty: false}); err != nil {
				return "", err
			}
		}
	}

	for _, c := range containers {
		if s.isPodInfraName(c.Name) {
			continue
		}
		if c.State == "running" {
			continue
		}
		if _, err := s.containerHandler.Start(container.ServiceStartModel{ContainerId: c.ContainerId, Tty: false}); err != nil {
			return "", err
		}
	}

	if err := s.psmHandler.UpdatePod(podId, "running"); err != nil {
		return "", err
	}

	return podId, nil
}
