package service

import (
	"errors"

	"github.com/wazadio/coupon-system/internal/models"
	"github.com/wazadio/coupon-system/internal/repository"
)

// CouponService defines the interface for coupon business logic
type CouponService interface {
	CreateCoupon(req *models.CreateCouponRequest) error
	ClaimCoupon(req *models.ClaimCouponRequest) error
	GetCouponDetails(name string) (*models.CouponDetailResponse, error)
	UpdateCoupon(name string) (rowsAffected int64, err error)
}

// couponService handles business logic for coupons
type couponService struct {
	repo repository.CouponRepository
}

// NewCouponService creates a new CouponService with injected repository
func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{
		repo: repo,
	}
}

// CreateCoupon creates a new coupon
func (s *couponService) CreateCoupon(req *models.CreateCouponRequest) error {
	// Validate input
	if req.Name == "" {
		return errors.New("coupon name is required")
	}
	if req.Amount <= 0 {
		return errors.New("coupon amount must be greater than 0")
	}

	return s.repo.CreateCoupon(req.Name, req.Amount)
}

// ClaimCoupon attempts to claim a coupon for a user
func (s *couponService) ClaimCoupon(req *models.ClaimCouponRequest) error {
	// Validate input
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.CouponName == "" {
		return errors.New("coupon_name is required")
	}

	return s.repo.ClaimCoupon(req.UserID, req.CouponName)
}

// GetCouponDetails retrieves coupon details with all claimed users
func (s *couponService) GetCouponDetails(name string) (*models.CouponDetailResponse, error) {
	if name == "" {
		return nil, errors.New("coupon name is required")
	}

	return s.repo.GetCouponByName(name)
}

func (s *couponService) UpdateCoupon(name string) (rowsAffected int64, err error) {
	if name == "" {
		return 0, errors.New("coupon name is required")
	}

	return s.repo.Update(name)
}
