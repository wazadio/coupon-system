package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wazadio/coupon-system/internal/models"
	"github.com/wazadio/coupon-system/internal/repository"
)

// MockCouponRepository is a mock implementation of CouponRepository
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) CreateCoupon(name string, amount int) error {
	args := m.Called(name, amount)
	return args.Error(0)
}

func (m *MockCouponRepository) ClaimCoupon(userID, couponName string) error {
	args := m.Called(userID, couponName)
	return args.Error(0)
}

func (m *MockCouponRepository) GetCouponByName(name string) (*models.CouponDetailResponse, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CouponDetailResponse), args.Error(1)
}

func (m *MockCouponRepository) Update(name string) (int64, error) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func TestCreateCoupon_Success(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	mockRepo.On("CreateCoupon", "FLASH25", 100).Return(nil)

	err := service.CreateCoupon(req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCoupon_EmptyName(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "",
		Amount: 100,
	}

	err := service.CreateCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "coupon name is required", err.Error())
}

func TestCreateCoupon_ZeroAmount(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 0,
	}

	err := service.CreateCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "coupon amount must be greater than 0", err.Error())
}

func TestCreateCoupon_NegativeAmount(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: -10,
	}

	err := service.CreateCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "coupon amount must be greater than 0", err.Error())
}

func TestCreateCoupon_AlreadyExists(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	mockRepo.On("CreateCoupon", "FLASH25", 100).Return(repository.ErrCouponAlreadyExists)

	err := service.CreateCoupon(req)
	assert.Equal(t, repository.ErrCouponAlreadyExists, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCoupon_RepositoryError(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	mockRepo.On("CreateCoupon", "FLASH25", 100).Return(errors.New("database error"))

	err := service.CreateCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestClaimCoupon_Success(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockRepo.On("ClaimCoupon", "user1", "FLASH25").Return(nil)

	err := service.ClaimCoupon(req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestClaimCoupon_EmptyUserID(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "",
		CouponName: "FLASH25",
	}

	err := service.ClaimCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "user_id is required", err.Error())
}

func TestClaimCoupon_EmptyCouponName(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "",
	}

	err := service.ClaimCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "coupon_name is required", err.Error())
}

func TestClaimCoupon_CouponNotFound(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "NONEXISTENT",
	}

	mockRepo.On("ClaimCoupon", "user1", "NONEXISTENT").Return(repository.ErrCouponNotFound)

	err := service.ClaimCoupon(req)
	assert.Equal(t, repository.ErrCouponNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestClaimCoupon_AlreadyClaimed(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockRepo.On("ClaimCoupon", "user1", "FLASH25").Return(repository.ErrAlreadyClaimed)

	err := service.ClaimCoupon(req)
	assert.Equal(t, repository.ErrAlreadyClaimed, err)
	mockRepo.AssertExpectations(t)
}

func TestClaimCoupon_NoStockAvailable(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockRepo.On("ClaimCoupon", "user1", "FLASH25").Return(repository.ErrNoStockAvailable)

	err := service.ClaimCoupon(req)
	assert.Equal(t, repository.ErrNoStockAvailable, err)
	mockRepo.AssertExpectations(t)
}

func TestClaimCoupon_RepositoryError(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	req := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockRepo.On("ClaimCoupon", "user1", "FLASH25").Return(errors.New("database error"))

	err := service.ClaimCoupon(req)
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetCouponDetails_Success(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	expectedResponse := &models.CouponDetailResponse{
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 75,
		ClaimedBy:       []string{},
	}

	mockRepo.On("GetCouponByName", "FLASH25").Return(expectedResponse, nil)

	result, err := service.GetCouponDetails("FLASH25")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FLASH25", result.Name)
	assert.Equal(t, 100, result.Amount)
	assert.Equal(t, 75, result.RemainingAmount)
	mockRepo.AssertExpectations(t)
}

func TestGetCouponDetails_EmptyName(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	result, err := service.GetCouponDetails("")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, "coupon name is required", err.Error())
}

func TestGetCouponDetails_NotFound(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	mockRepo.On("GetCouponByName", "NONEXISTENT").Return(nil, repository.ErrCouponNotFound)

	result, err := service.GetCouponDetails("NONEXISTENT")
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrCouponNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestGetCouponDetails_RepositoryError(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	mockRepo.On("GetCouponByName", "FLASH25").Return(nil, errors.New("database error"))

	result, err := service.GetCouponDetails("FLASH25")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUpdateCoupon_Success(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	mockRepo.On("Update", "FLASH25").Return(int64(1), nil)

	rowsAffected, err := service.UpdateCoupon("FLASH25")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	mockRepo.AssertExpectations(t)
}

func TestUpdateCoupon_EmptyName(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	rowsAffected, err := service.UpdateCoupon("")
	assert.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected)
	assert.Equal(t, "coupon name is required", err.Error())
}

func TestUpdateCoupon_RepositoryError(t *testing.T) {
	mockRepo := new(MockCouponRepository)
	service := NewCouponService(mockRepo)

	mockRepo.On("Update", "FLASH25").Return(int64(0), errors.New("database error"))

	rowsAffected, err := service.UpdateCoupon("FLASH25")
	assert.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}
