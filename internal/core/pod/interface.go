package pod

type PodServiceHandler interface {
	Run(runParameter ServiceRunModel) (string, error)
	Stop(podId string) (string, error)
	Remove(podId string) (string, error)
	GetPodList() ([]PodState, error)
	GetPodById(podId string) (PodState, error)
}
