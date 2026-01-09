package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateCoupon_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectExec("INSERT INTO coupons").
		WithArgs("FLASH25", 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateCoupon("FLASH25", 100)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCoupon_DuplicateCoupon(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	pqErr := &pq.Error{Code: "23505"}
	mock.ExpectExec("INSERT INTO coupons").
		WithArgs("FLASH25", 100).
		WillReturnError(pqErr)

	err = repo.CreateCoupon("FLASH25", 100)
	assert.Equal(t, ErrCouponAlreadyExists, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCoupon_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectExec("INSERT INTO coupons").
		WithArgs("FLASH25", 100).
		WillReturnError(errors.New("database connection lost"))

	err = repo.CreateCoupon("FLASH25", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating coupon")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(10))
	mock.ExpectExec("INSERT INTO claims").
		WithArgs("user1", "FLASH25").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE coupons SET remaining_amount").
		WithArgs("FLASH25").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Note: The actual implementation has time.Sleep(2 * time.Second) which we can't mock
	// For testing purposes, you may want to refactor the repository to inject a sleep function
	// For now, this test will take 2 seconds due to the sleep in the actual code

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_CouponNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("NONEXISTENT").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "NONEXISTENT")
	assert.Equal(t, ErrCouponNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_NoStockAvailable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(0))
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Equal(t, ErrNoStockAvailable, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_AlreadyClaimed(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	pqErr := &pq.Error{Code: "23505"}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(10))
	mock.ExpectExec("INSERT INTO claims").
		WithArgs("user1", "FLASH25").
		WillReturnError(pqErr)
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Equal(t, ErrAlreadyClaimed, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_TransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin().WillReturnError(errors.New("connection pool exhausted"))

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error starting transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_SelectError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnError(errors.New("connection timeout"))
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error checking coupon")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_InsertClaimError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(10))
	mock.ExpectExec("INSERT INTO claims").
		WithArgs("user1", "FLASH25").
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating claim")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_UpdateCouponError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(10))
	mock.ExpectExec("INSERT INTO claims").
		WithArgs("user1", "FLASH25").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE coupons SET remaining_amount").
		WithArgs("FLASH25").
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error updating coupon stock")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimCoupon_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &couponRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT remaining_amount FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(sqlmock.NewRows([]string{"remaining_amount"}).AddRow(10))
	mock.ExpectExec("INSERT INTO claims").
		WithArgs("user1", "FLASH25").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE coupons SET remaining_amount").
		WithArgs("FLASH25").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err = repo.ClaimCoupon("user1", "FLASH25")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error committing transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouponByName_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	now := time.Now()
	couponRows := sqlmock.NewRows([]string{"id", "name", "amount", "remaining_amount", "created_at", "updated_at"}).
		AddRow(1, "FLASH25", 100, 75, now, now)

	claimRows := sqlmock.NewRows([]string{"user_id"}).
		AddRow("user1").
		AddRow("user2")

	mock.ExpectQuery("SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(couponRows)

	mock.ExpectQuery("SELECT user_id FROM claims WHERE coupon_name").
		WithArgs("FLASH25").
		WillReturnRows(claimRows)

	result, err := repo.GetCouponByName("FLASH25")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FLASH25", result.Name)
	assert.Equal(t, 100, result.Amount)
	assert.Equal(t, 75, result.RemainingAmount)
	assert.Len(t, result.ClaimedBy, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouponByName_SuccessNoClaims(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	now := time.Now()
	couponRows := sqlmock.NewRows([]string{"id", "name", "amount", "remaining_amount", "created_at", "updated_at"}).
		AddRow(1, "FLASH25", 100, 100, now, now)

	claimRows := sqlmock.NewRows([]string{"user_id"})

	mock.ExpectQuery("SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(couponRows)

	mock.ExpectQuery("SELECT user_id FROM claims WHERE coupon_name").
		WithArgs("FLASH25").
		WillReturnRows(claimRows)

	result, err := repo.GetCouponByName("FLASH25")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FLASH25", result.Name)
	assert.Len(t, result.ClaimedBy, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouponByName_CouponNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectQuery("SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name").
		WithArgs("NONEXISTENT").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetCouponByName("NONEXISTENT")
	assert.Nil(t, result)
	assert.Equal(t, ErrCouponNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouponByName_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectQuery("SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnError(errors.New("connection timeout"))

	result, err := repo.GetCouponByName("FLASH25")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting coupon")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCouponByName_ClaimsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	now := time.Now()
	couponRows := sqlmock.NewRows([]string{"id", "name", "amount", "remaining_amount", "created_at", "updated_at"}).
		AddRow(1, "FLASH25", 100, 75, now, now)

	mock.ExpectQuery("SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name").
		WithArgs("FLASH25").
		WillReturnRows(couponRows)

	mock.ExpectQuery("SELECT user_id FROM claims WHERE coupon_name").
		WithArgs("FLASH25").
		WillReturnError(errors.New("connection timeout"))

	result, err := repo.GetCouponByName("FLASH25")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting claims")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectExec("UPDATE coupons SET updated_at").
		WithArgs("FLASH25").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rowsAffected, err := repo.Update("FLASH25")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_NoRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectExec("UPDATE coupons SET updated_at").
		WithArgs("NONEXISTENT").
		WillReturnResult(sqlmock.NewResult(0, 0))

	rowsAffected, err := repo.Update("NONEXISTENT")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewCouponRepository(db)

	mock.ExpectExec("UPDATE coupons SET updated_at").
		WithArgs("FLASH25").
		WillReturnError(errors.New("database error"))

	rowsAffected, err := repo.Update("FLASH25")
	assert.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
