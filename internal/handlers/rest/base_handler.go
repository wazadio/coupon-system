package rest

import (
	"net/http"

	"github.com/wazadio/coupon-system/pkg/rest"
)

type BaseHandler struct{}

// healthCheck endpoint to verify the service is running
func (h *BaseHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	rest.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
