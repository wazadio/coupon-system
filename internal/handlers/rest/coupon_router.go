package rest

import (
	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the HTTP router with injected dependencies
func (h *CouponHandler) SetupRouter(router *mux.Router) {
	// API routes
	api := router.PathPrefix("/coupons").Subrouter()

	// Coupon routes
	api.HandleFunc("", h.CreateCoupon).Methods("POST")
	api.HandleFunc("/claim", h.ClaimCoupon).Methods("POST")
	api.HandleFunc("/{name}", h.GetCouponDetails).Methods("GET")
	api.HandleFunc("/{name}", h.UpdateCoupon).Methods("PUT", "PATCH")
}
