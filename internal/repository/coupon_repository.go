package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/wazadio/coupon-system/internal/models"
)

var (
	ErrCouponNotFound      = errors.New("coupon not found")
	ErrCouponAlreadyExists = errors.New("coupon already exists")
	ErrAlreadyClaimed      = errors.New("user already claimed this coupon")
	ErrNoStockAvailable    = errors.New("no stock available")
)

// CouponRepository defines the interface for coupon data operations
type CouponRepository interface {
	CreateCoupon(name string, amount int) error
	ClaimCoupon(userID, couponName string) error
	GetCouponByName(name string) (*models.CouponDetailResponse, error)
	Update(name string) (rowsAffected int64, err error)
}

// couponRepository handles database operations for coupons
type couponRepository struct {
	db *sql.DB
}

// NewCouponRepository creates a new CouponRepository with injected database connection
func NewCouponRepository(db *sql.DB) CouponRepository {
	return &couponRepository{
		db: db,
	}
}

// CreateCoupon creates a new coupon
func (r *couponRepository) CreateCoupon(name string, amount int) error {
	query := `
		INSERT INTO coupons (name, amount, remaining_amount)
		VALUES ($1, $2, $2)
	`

	_, err := r.db.Exec(query, name, amount)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrCouponAlreadyExists
		}
		return fmt.Errorf("error creating coupon: %v", err)
	}

	return nil
}

// ClaimCoupon attempts to claim a coupon for a user with proper transaction handling
func (r *couponRepository) ClaimCoupon(userID, couponName string) error {
	// Start a transaction with default READ COMMITTED isolation level
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Lock the coupon row for update to prevent race conditions
	// SELECT FOR UPDATE causes other transactions to wait (not fail)
	var remainingAmount int
	query := `
		SELECT remaining_amount 
		FROM coupons 
		WHERE name = $1 
		FOR UPDATE
	`
	err = tx.QueryRow(query, couponName).Scan(&remainingAmount)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCouponNotFound
		}
		return fmt.Errorf("error checking coupon: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Check if stock is available
	if remainingAmount <= 0 {
		return ErrNoStockAvailable
	}

	// Try to insert claim record
	// This will fail if the user already claimed this coupon (unique constraint)
	insertQuery := `
		INSERT INTO claims (user_id, coupon_name)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(insertQuery, userID, couponName)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAlreadyClaimed
		}
		return fmt.Errorf("error creating claim: %v", err)
	}

	// Decrement the coupon stock
	updateQuery := `
		UPDATE coupons 
		SET remaining_amount = remaining_amount - 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE name = $1
	`
	_, err = tx.Exec(updateQuery, couponName)
	if err != nil {
		return fmt.Errorf("error updating coupon stock: %v", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// GetCouponByName retrieves a coupon by name with all users who claimed it
func (r *couponRepository) GetCouponByName(name string) (*models.CouponDetailResponse, error) {
	// Get coupon details
	var coupon models.Coupon
	query := `
		SELECT id, name, amount, remaining_amount, created_at, updated_at
		FROM coupons
		WHERE name = $1
	`
	err := r.db.QueryRow(query, name).Scan(
		&coupon.ID,
		&coupon.Name,
		&coupon.Amount,
		&coupon.RemainingAmount,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCouponNotFound
		}
		return nil, fmt.Errorf("error getting coupon: %v", err)
	}

	// Get all users who claimed this coupon
	claimsQuery := `
		SELECT user_id
		FROM claims
		WHERE coupon_name = $1
		ORDER BY claimed_at ASC
	`
	rows, err := r.db.Query(claimsQuery, name)
	if err != nil {
		return nil, fmt.Errorf("error getting claims: %v", err)
	}
	defer rows.Close()

	claimedBy := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("error scanning claim: %v", err)
		}
		claimedBy = append(claimedBy, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating claims: %v", err)
	}

	response := &models.CouponDetailResponse{
		Name:            coupon.Name,
		Amount:          coupon.Amount,
		RemainingAmount: coupon.RemainingAmount,
		ClaimedBy:       claimedBy,
	}

	return response, nil
}

func (r *couponRepository) Update(name string) (rowsAffected int64, err error) {
	updateQuery := `
		UPDATE coupons 
		SET updated_at = NOW()
		WHERE name = $1;
	`

	result, err := r.db.Exec(updateQuery, name)
	if err != nil {
		return
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return
	}

	return
}
