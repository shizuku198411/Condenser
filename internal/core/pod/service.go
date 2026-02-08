package pod

import (
	"condenser/internal/store/psm"
	"condenser/internal/utils"
)

func NewPodService() *PodService {
	return &PodService{
		psmHandler: psm.NewPsmManager(psm.NewPsmStore(utils.PsmStorePath)),
	}
}

type PodService struct {
	psmHandler psm.PsmHandler
}
