package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCouponRequest_Valid(t *testing.T) {
	req := &CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	assert.Equal(t, "FLASH25", req.Name)
	assert.Equal(t, 100, req.Amount)
}

func TestCreateCouponRequest_Empty(t *testing.T) {
	req := &CreateCouponRequest{}

	assert.Empty(t, req.Name)
	assert.Equal(t, 0, req.Amount)
}

func TestClaimCouponRequest_Valid(t *testing.T) {
	req := &ClaimCouponRequest{
		UserID:     "user123",
		CouponName: "FLASH25",
	}

	assert.Equal(t, "user123", req.UserID)
	assert.Equal(t, "FLASH25", req.CouponName)
}

func TestClaimCouponRequest_Empty(t *testing.T) {
	req := &ClaimCouponRequest{}

	assert.Empty(t, req.UserID)
	assert.Empty(t, req.CouponName)
}

func TestCoupon_Initialization(t *testing.T) {
	now := time.Now()
	coupon := &Coupon{
		ID:              1,
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 75,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, int64(1), coupon.ID)
	assert.Equal(t, "FLASH25", coupon.Name)
	assert.Equal(t, 100, coupon.Amount)
	assert.Equal(t, 75, coupon.RemainingAmount)
	assert.Equal(t, now, coupon.CreatedAt)
	assert.Equal(t, now, coupon.UpdatedAt)
}

func TestCoupon_EmptyStruct(t *testing.T) {
	coupon := &Coupon{}

	assert.Equal(t, int64(0), coupon.ID)
	assert.Empty(t, coupon.Name)
	assert.Equal(t, 0, coupon.Amount)
	assert.Equal(t, 0, coupon.RemainingAmount)
	assert.True(t, coupon.CreatedAt.IsZero())
	assert.True(t, coupon.UpdatedAt.IsZero())
}



func TestCouponDetailResponse_WithUsers(t *testing.T) {
	response := &CouponDetailResponse{
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 75,
		ClaimedBy:       []string{"user1", "user2"},
	}

	assert.Equal(t, "FLASH25", response.Name)
	assert.Equal(t, 100, response.Amount)
	assert.Equal(t, 75, response.RemainingAmount)
	assert.Len(t, response.ClaimedBy, 2)
	assert.Equal(t, "user1", response.ClaimedBy[0])
	assert.Equal(t, "user2", response.ClaimedBy[1])
}

func TestCouponDetailResponse_NoUsers(t *testing.T) {
	response := &CouponDetailResponse{
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 100,
		ClaimedBy:       []string{},
	}

	assert.Equal(t, "FLASH25", response.Name)
	assert.Equal(t, 100, response.Amount)
	assert.Equal(t, 100, response.RemainingAmount)
	assert.Len(t, response.ClaimedBy, 0)
}

func TestCouponDetailResponse_NilUsers(t *testing.T) {
	response := &CouponDetailResponse{
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 100,
		ClaimedBy:       nil,
	}

	assert.Equal(t, "FLASH25", response.Name)
	assert.Equal(t, 100, response.Amount)
	assert.Equal(t, 100, response.RemainingAmount)
	assert.Nil(t, response.ClaimedBy)
}

func TestCouponDetailResponse_Empty(t *testing.T) {
	response := &CouponDetailResponse{}

	assert.Empty(t, response.Name)
	assert.Equal(t, 0, response.Amount)
	assert.Equal(t, 0, response.RemainingAmount)
	assert.Nil(t, response.ClaimedBy)
}

func TestCoupon_FullClaimed(t *testing.T) {
	coupon := &Coupon{
		ID:              1,
		Name:            "FLASH25",
		Amount:          10,
		RemainingAmount: 0,
	}

	assert.Equal(t, 0, coupon.RemainingAmount)
	assert.Equal(t, 10, coupon.Amount)
}

func TestCoupon_PartiallyClaimed(t *testing.T) {
	coupon := &Coupon{
		ID:              1,
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 50,
	}

	assert.Equal(t, 50, coupon.RemainingAmount)
	assert.Equal(t, 100, coupon.Amount)
}

func TestCreateCouponRequest_NegativeAmount(t *testing.T) {
	req := &CreateCouponRequest{
		Name:   "FLASH25",
		Amount: -10,
	}

	assert.Equal(t, "FLASH25", req.Name)
	assert.Equal(t, -10, req.Amount)
	// Validation should be done at service layer, model allows any value
}

func TestCreateCouponRequest_ZeroAmount(t *testing.T) {
	req := &CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 0,
	}

	assert.Equal(t, "FLASH25", req.Name)
	assert.Equal(t, 0, req.Amount)
	// Validation should be done at service layer, model allows any value
}


