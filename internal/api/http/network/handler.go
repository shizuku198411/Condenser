package network

import (
	apimodel "condenser/internal/api/http/utils"
	"condenser/internal/core/network"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{
		serviceHandler: network.NewNetworkService(),
	}
}

type RequestHandler struct {
	serviceHandler network.NetworkServiceHandler
}

// CreateBridge godoc
// @Summary create bridge
// @Description create new bridge
// @Tags Network
// @Accept json
// @Produce json
// @Param request body CreateBridgeRequest true "Bridge Information"
// @Success 201 {object} apimodel.ApiResponse
// @Router /v1/networks [post]
func (h *RequestHandler) CreateBridge(w http.ResponseWriter, r *http.Request) {
	// decode request
	var req CreateBridgeRequest
	if err := apimodel.DecodeRequestBody(r, &req); err != nil {
		apimodel.RespondFail(w, http.StatusBadRequest, "invalid json: "+err.Error(), nil)
	}

	// service
	if err := h.serviceHandler.CreateNewNetwork(
		network.ServiceNewNetworkModel{
			Bridge: req.Bridge,
		},
	); err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "create failed: "+err.Error(), nil)
		return
	}

	// encode response
	apimodel.RespondSuccess(w, http.StatusOK, "create network: "+req.Bridge+" completed", req)
}

// DeleteBridge godoc
// @Summary delete bridge
// @Description delete new bridge
// @Tags Network
// @Accept json
// @Produce json
// @Success 201 {object} apimodel.ApiResponse
// @Router /v1/networks/{bridge}/actions/delete [delete]
func (h *RequestHandler) DeleteBridge(w http.ResponseWriter, r *http.Request) {
	bridge := chi.URLParam(r, "bridge")
	if bridge == "" {
		apimodel.RespondFail(w, http.StatusBadRequest, "missing brige", nil)
		return
	}

	// service
	if err := h.serviceHandler.RemoveNetwork(
		network.ServiceRemoveNetworkModel{
			Bridge: bridge,
		},
	); err != nil {
		apimodel.RespondFail(w, http.StatusInternalServerError, "delete failed: "+err.Error(), nil)
		return
	}

	// encode response
	apimodel.RespondSuccess(w, http.StatusOK, "delete network: "+bridge+" completed", nil)
}
