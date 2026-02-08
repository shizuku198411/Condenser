package pod

import (
	"fmt"

	"condenser/internal/utils"
)

// == service: run pod sandbox ==
func (s *PodService) Run(runParameter ServiceRunModel) (string, error) {
	if runParameter.Name == "" {
		return "", fmt.Errorf("pod name is required")
	}
	if runParameter.Namespace == "" {
		return "", fmt.Errorf("pod namespace is required")
	}
	if s.psmHandler.IsNameAlreadyUsed(runParameter.Name, runParameter.Namespace) {
		return "", fmt.Errorf("pod name already used: %s/%s", runParameter.Namespace, runParameter.Name)
	}

	podId := utils.NewUlid()
	if err := s.psmHandler.StorePod(
		podId,
		runParameter.Name,
		runParameter.Namespace,
		runParameter.UID,
		"created",
		runParameter.NetworkNS,
		runParameter.IPCNS,
		runParameter.UTSNS,
		runParameter.UserNS,
		runParameter.Labels,
		runParameter.Annotations,
	); err != nil {
		return "", err
	}

	return podId, nil
}
