package rest

import (
	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the HTTP router with injected dependencies
func (h *BaseHandler) SetupRouter(router *mux.Router) {
	// Use /health to work with PathPrefix subrouter
	router.HandleFunc("/health", h.healthCheck).Methods("GET")
}
