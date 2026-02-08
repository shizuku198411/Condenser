package pod

// == service: remove pod sandbox ==
func (s *PodService) Remove(podId string) (string, error) {
	if err := s.psmHandler.RemovePod(podId); err != nil {
		return "", err
	}
	return podId, nil
}
