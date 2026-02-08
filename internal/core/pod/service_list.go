package pod

// == service: list pods ==
func (s *PodService) GetPodList() ([]PodState, error) {
	podList, err := s.psmHandler.GetPodList()
	if err != nil {
		return nil, err
	}
	var result []PodState
	for _, p := range podList {
		result = append(result, PodState{
			PodId:       p.PodId,
			Name:        p.Name,
			Namespace:   p.Namespace,
			UID:         p.UID,
			State:       p.State,
			NetworkNS:   p.NetworkNS,
			IPCNS:       p.IPCNS,
			UTSNS:       p.UTSNS,
			UserNS:      p.UserNS,
			Labels:      p.Labels,
			Annotations: p.Annotations,
			CreatedAt:   p.CreatedAt,
			StartedAt:   p.StartedAt,
			StoppedAt:   p.StoppedAt,
		})
	}
	return result, nil
}

// == service: get pod by id ==
func (s *PodService) GetPodById(podId string) (PodState, error) {
	p, err := s.psmHandler.GetPodById(podId)
	if err != nil {
		return PodState{}, err
	}
	return PodState{
		PodId:       p.PodId,
		Name:        p.Name,
		Namespace:   p.Namespace,
		UID:         p.UID,
		State:       p.State,
		NetworkNS:   p.NetworkNS,
		IPCNS:       p.IPCNS,
		UTSNS:       p.UTSNS,
		UserNS:      p.UserNS,
		Labels:      p.Labels,
		Annotations: p.Annotations,
		CreatedAt:   p.CreatedAt,
		StartedAt:   p.StartedAt,
		StoppedAt:   p.StoppedAt,
	}, nil
}
