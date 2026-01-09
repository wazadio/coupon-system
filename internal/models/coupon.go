package models

import "time"

// Coupon represents a coupon in the system
type Coupon struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Amount          int       `json:"amount"`
	RemainingAmount int       `json:"remaining_amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Claim represents a user's claim of a coupon
type Claim struct {
	ID         int       `json:"id"`
	UserID     string    `json:"user_id"`
	CouponName string    `json:"coupon_name"`
	ClaimedAt  time.Time `json:"claimed_at"`
}

// CreateCouponRequest is the request body for creating a coupon
type CreateCouponRequest struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

// ClaimCouponRequest is the request body for claiming a coupon
type ClaimCouponRequest struct {
	UserID     string `json:"user_id"`
	CouponName string `json:"coupon_name"`
}

// CouponDetailResponse is the response for getting coupon details
type CouponDetailResponse struct {
	Name            string   `json:"name"`
	Amount          int      `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}
