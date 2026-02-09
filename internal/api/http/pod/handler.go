package pod

import (
	"net/http"

	apimodel "condenser/internal/api/http/utils"
	"condenser/internal/core/pod"

	"github.com/go-chi/chi/v5"
)

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{
		serviceHandler: pod.NewPodService(),
	}
}

type RequestHandler struct {
	serviceHandler pod.PodServiceHandler
}

// CreatePod godoc
// @Summary create pod sandbox
// @Description create a pod sandbox (no container start)
// @Tags pods
// @Accept json
// @Produce json
// @Param request body CreatePodRequest true "Pod Spec"
// @Success 201 {object} apimodel.ApiResponse
// @Router /v1/pods [post]
func (h *RequestHandler) CreatePod(w http.ResponseWriter, r *http.Request) {
	var req CreatePodRequest
	if err := apimodel.DecodeRequestBody(r, &req); err != nil {
		apimodel.RespondFail(w, http.StatusBadRequest, "invalid json: "+err.Error(), CreatePodResponse{PodId: ""})
		return
	}

	podId, err := h.serviceHandler.Create(pod.ServiceCreateModel{
		Name:        req.Name,
		Namespace:   req.Namespace,
		UID:         req.UID,
		NetworkNS:   req.NetworkNS,
		IPCNS:       req.IPCNS,
		UTSNS:       req.UTSNS,
		UserNS:      req.UserNS,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	})
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "create pod failed: "+err.Error(), CreatePodResponse{PodId: ""})
		return
	}

	apimodel.RespondSuccess(w, http.StatusOK, "pod created", CreatePodResponse{PodId: podId})
}

// StartPod godoc
// @Summary start pod sandbox
// @Description start a pod sandbox
// @Tags pods
// @Param podId path string true "Pod ID"
// @Success 200 {object} apimodel.ApiResponse
// @Router /v1/pods/{podId}/actions/start [post]
func (h *RequestHandler) StartPod(w http.ResponseWriter, r *http.Request) {
	podId := chi.URLParam(r, "podId")
	if podId == "" {
		apimodel.RespondFail(w, http.StatusBadRequest, "missing podId", StartPodResponse{PodId: ""})
		return
	}

	result, err := h.serviceHandler.Start(podId)
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "start pod failed: "+err.Error(), StartPodResponse{PodId: podId})
		return
	}

	apimodel.RespondSuccess(w, http.StatusOK, "pod started", StartPodResponse{PodId: result})
}

// StopPod godoc
// @Summary stop pod sandbox
// @Description stop a pod sandbox
// @Tags pods
// @Param podId path string true "Pod ID"
// @Success 200 {object} apimodel.ApiResponse
// @Router /v1/pods/{podId}/actions/stop [post]
func (h *RequestHandler) StopPod(w http.ResponseWriter, r *http.Request) {
	podId := chi.URLParam(r, "podId")
	if podId == "" {
		apimodel.RespondFail(w, http.StatusBadRequest, "missing podId", StopPodResponse{PodId: ""})
		return
	}

	result, err := h.serviceHandler.Stop(podId)
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "stop pod failed: "+err.Error(), StopPodResponse{PodId: podId})
		return
	}

	apimodel.RespondSuccess(w, http.StatusOK, "pod stopped", StopPodResponse{PodId: result})
}

// RemovePod godoc
// @Summary remove pod sandbox
// @Description remove a pod sandbox
// @Tags pods
// @Param podId path string true "Pod ID"
// @Success 200 {object} apimodel.ApiResponse
// @Router /v1/pods/{podId} [delete]
func (h *RequestHandler) RemovePod(w http.ResponseWriter, r *http.Request) {
	podId := chi.URLParam(r, "podId")
	if podId == "" {
		apimodel.RespondFail(w, http.StatusBadRequest, "missing podId", RemovePodResponse{PodId: ""})
		return
	}

	result, err := h.serviceHandler.Remove(podId)
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "remove pod failed: "+err.Error(), RemovePodResponse{PodId: podId})
		return
	}

	apimodel.RespondSuccess(w, http.StatusOK, "pod removed", RemovePodResponse{PodId: result})
}

// GetPodList godoc
// @Summary list pods
// @Description list pod sandbox
// @Tags pods
// @Success 200 {object} apimodel.ApiResponse
// @Router /v1/pods [get]
func (h *RequestHandler) GetPodList(w http.ResponseWriter, r *http.Request) {
	podList, err := h.serviceHandler.GetPodList()
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "retrieve pod list failed: "+err.Error(), nil)
		return
	}
	apimodel.RespondSuccess(w, http.StatusOK, "retrieve pod list success", podList)
}

// GetPodById godoc
// @Summary get pod detail
// @Description get pod sandbox detail
// @Tags pods
// @Param podId path string true "Pod ID"
// @Success 200 {object} apimodel.ApiResponse
// @Router /v1/pods/{podId} [get]
func (h *RequestHandler) GetPodById(w http.ResponseWriter, r *http.Request) {
	podId := chi.URLParam(r, "podId")
	if podId == "" {
		apimodel.RespondFail(w, http.StatusBadRequest, "missing podId", nil)
		return
	}

	podInfo, err := h.serviceHandler.GetPodById(podId)
	if err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "retrieve pod failed: "+err.Error(), nil)
		return
	}

	apimodel.RespondSuccess(w, http.StatusOK, "retrieve pod success", podInfo)
}
