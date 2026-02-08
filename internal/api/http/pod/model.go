package pod

type RunPodRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	UID         string            `json:"uid"`
	NetworkNS   string            `json:"networkNS"`
	IPCNS       string            `json:"ipcNS"`
	UTSNS       string            `json:"utsNS"`
	UserNS      string            `json:"userNS"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type RunPodResponse struct {
	PodId string `json:"podId"`
}

type StopPodResponse struct {
	PodId string `json:"podId"`
}

type RemovePodResponse struct {
	PodId string `json:"podId"`
}
