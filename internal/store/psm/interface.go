package psm

type PsmStoreHandler interface {
	SetPodState() error
}

type PsmHandler interface {
	StorePod(podId, name, namespace, uid, state, networkNS, ipcNS, utsNS, userNS string, labels, annotations map[string]string) error
	RemovePod(podId string) error
	UpdatePod(podId string, state string) error
	UpdatePodNamespaces(ownerPid int, podId, networkNS, ipcNS, utsNS, userNS string) error
	GetPodList() ([]PodInfo, error)
	GetPodById(podId string) (PodInfo, error)
	IsNameAlreadyUsed(name, namespace string) bool
	GetPodIdByName(name, namespace string) (string, error)
	ResolvePodId(str, namespace string) (string, error)
	IsPodExist(podId string) bool
	IsPodOwner(podId string) bool
	GetPodOwnerPid(podId string) (int, error)
}
