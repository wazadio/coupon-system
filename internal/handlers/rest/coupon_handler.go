package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wazadio/coupon-system/internal/models"
	"github.com/wazadio/coupon-system/internal/repository"
	"github.com/wazadio/coupon-system/internal/service"
	"github.com/wazadio/coupon-system/pkg/logger"
	pkgRest "github.com/wazadio/coupon-system/pkg/rest"
)

// CouponHandler handles HTTP requests for coupons
type CouponHandler struct {
	service service.CouponService
}

// NewCouponHandler creates a new CouponHandler with injected service
func NewCouponHandler(service service.CouponService) *CouponHandler {
	return &CouponHandler{
		service: service,
	}
}

// CreateCoupon handles POST /api/coupons
func (h *CouponHandler) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCouponRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Print(r.Context(), logger.LevelError, err.Error())
		pkgRest.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create coupon
	err := h.service.CreateCoupon(&req)
	if err != nil {
		if err == repository.ErrCouponAlreadyExists {
			logger.Print(r.Context(), logger.LevelError, "Coupon already exists")
			pkgRest.RespondWithError(w, http.StatusConflict, "Coupon already exists")
			return
		}

		logger.Print(r.Context(), logger.LevelError, err.Error())
		pkgRest.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return 201 Created
	pkgRest.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Coupon created successfully"})
}

// ClaimCoupon handles POST /api/coupons/claim
func (h *CouponHandler) ClaimCoupon(w http.ResponseWriter, r *http.Request) {
	var req models.ClaimCouponRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Print(r.Context(), logger.LevelError, err.Error())
		pkgRest.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Attempt to claim coupon
	err := h.service.ClaimCoupon(&req)
	if err != nil {
		switch err {
		case repository.ErrAlreadyClaimed:
			logger.Print(r.Context(), logger.LevelError, "User already claimed this coupon")
			pkgRest.RespondWithError(w, http.StatusConflict, "User already claimed this coupon")
			return
		case repository.ErrNoStockAvailable:
			logger.Print(r.Context(), logger.LevelError, "No stock available for this coupon")
			pkgRest.RespondWithError(w, http.StatusBadRequest, "No stock available")
			return
		case repository.ErrCouponNotFound:
			logger.Print(r.Context(), logger.LevelError, "Coupon not found")
			pkgRest.RespondWithError(w, http.StatusNotFound, "Coupon not found")
			return
		default:
			logger.Print(r.Context(), logger.LevelError, err.Error())
			pkgRest.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Return 200 OK
	pkgRest.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Coupon claimed successfully"})
}

// GetCouponDetails handles GET /api/coupons/{name}
func (h *CouponHandler) GetCouponDetails(w http.ResponseWriter, r *http.Request) {
	// Get coupon name from URL parameter
	vars := mux.Vars(r)
	name := vars["name"]

	// Get coupon details
	details, err := h.service.GetCouponDetails(name)
	if err != nil {
		if err == repository.ErrCouponNotFound {
			logger.Print(r.Context(), logger.LevelError, "Coupon not found")
			pkgRest.RespondWithError(w, http.StatusNotFound, "Coupon not found")
			return
		}
		logger.Print(r.Context(), logger.LevelError, err.Error())
		pkgRest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return coupon details
	pkgRest.RespondWithJSON(w, http.StatusOK, details)
}

func (h *CouponHandler) UpdateCoupon(w http.ResponseWriter, r *http.Request) {
	// Get coupon name from URL parameter
	vars := mux.Vars(r)
	name := vars["name"]

	// Update coupon
	rowsAffected, err := h.service.UpdateCoupon(name)
	if err != nil {
		if err == repository.ErrCouponNotFound {
			logger.Print(r.Context(), logger.LevelError, "Coupon not found")
			pkgRest.RespondWithError(w, http.StatusNotFound, "Coupon not found")
			return
		}
		logger.Print(r.Context(), logger.LevelError, err.Error())
		pkgRest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return update result
	pkgRest.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Coupon updated successfully",
		"rows_affected": rowsAffected,
	})
}
