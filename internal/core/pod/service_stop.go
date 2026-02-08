package pod

// == service: stop pod sandbox ==
func (s *PodService) Stop(podId string) (string, error) {
	if err := s.psmHandler.UpdatePod(podId, "stopped"); err != nil {
		return "", err
	}
	return podId, nil
}
