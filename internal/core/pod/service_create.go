package pod

import (
	"fmt"

	"condenser/internal/utils"
)

// == service: create pod sandbox ==
func (s *PodService) Create(createParameter ServiceCreateModel) (string, error) {
	if createParameter.Name == "" {
		return "", fmt.Errorf("pod name is required")
	}
	if createParameter.Namespace == "" {
		return "", fmt.Errorf("pod namespace is required")
	}
	if createParameter.UID == "" {
		createParameter.UID = utils.NewUlid()
	}

	if existingId, err := s.psmHandler.GetPodIdByName(createParameter.Name, createParameter.Namespace); err == nil {
		existing, err := s.psmHandler.GetPodById(existingId)
		if err == nil {
			if existing.UID != "" && createParameter.UID != "" && existing.UID == createParameter.UID {
				return existingId, nil
			}
			if existing.UID != "" && createParameter.UID != "" && existing.UID != createParameter.UID {
				if _, err := s.Remove(existingId); err != nil {
					return "", err
				}
			} else {
				return "", fmt.Errorf("pod name already used: %s/%s", createParameter.Namespace, createParameter.Name)
			}
		}
	}

	podId := utils.NewUlid()
	if err := s.psmHandler.StorePod(
		podId,
		createParameter.Name,
		createParameter.Namespace,
		createParameter.UID,
		"created",
		createParameter.NetworkNS,
		createParameter.IPCNS,
		createParameter.UTSNS,
		createParameter.UserNS,
		createParameter.Labels,
		createParameter.Annotations,
	); err != nil {
		return "", err
	}

	return podId, nil
}
