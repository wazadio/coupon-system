package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wazadio/coupon-system/internal/models"
	"github.com/wazadio/coupon-system/internal/repository"
	"github.com/wazadio/coupon-system/pkg/logger"
)

// MockCouponService is a mock implementation of CouponService
type MockCouponService struct {
	mock.Mock
}

func (m *MockCouponService) CreateCoupon(req *models.CreateCouponRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockCouponService) ClaimCoupon(req *models.ClaimCouponRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockCouponService) GetCouponDetails(name string) (*models.CouponDetailResponse, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CouponDetailResponse), args.Error(1)
}

func (m *MockCouponService) UpdateCoupon(name string) (int64, error) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func TestCreateCoupon_Handler_Success(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	mockService.On("CreateCoupon", reqBody).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.CreateCoupon(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon created successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestCreateCoupon_Handler_InvalidJSON(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.CreateCoupon(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestCreateCoupon_Handler_AlreadyExists(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.CreateCouponRequest{
		Name:   "FLASH25",
		Amount: 100,
	}

	mockService.On("CreateCoupon", reqBody).Return(repository.ErrCouponAlreadyExists)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.CreateCoupon(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon already exists", response["error"])

	mockService.AssertExpectations(t)
}

func TestCreateCoupon_Handler_ValidationError(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.CreateCouponRequest{
		Name:   "",
		Amount: 100,
	}

	mockService.On("CreateCoupon", reqBody).Return(errors.New("coupon name is required"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.CreateCoupon(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "coupon name is required", response["error"])

	mockService.AssertExpectations(t)
}

func TestClaimCoupon_Handler_Success(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockService.On("ClaimCoupon", reqBody).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon claimed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestClaimCoupon_Handler_InvalidJSON(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestClaimCoupon_Handler_AlreadyClaimed(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockService.On("ClaimCoupon", reqBody).Return(repository.ErrAlreadyClaimed)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "User already claimed this coupon", response["error"])

	mockService.AssertExpectations(t)
}

func TestClaimCoupon_Handler_NoStockAvailable(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "FLASH25",
	}

	mockService.On("ClaimCoupon", reqBody).Return(repository.ErrNoStockAvailable)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "No stock available", response["error"])

	mockService.AssertExpectations(t)
}

func TestClaimCoupon_Handler_CouponNotFound(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.ClaimCouponRequest{
		UserID:     "user1",
		CouponName: "NONEXISTENT",
	}

	mockService.On("ClaimCoupon", reqBody).Return(repository.ErrCouponNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestClaimCoupon_Handler_ValidationError(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	reqBody := &models.ClaimCouponRequest{
		UserID:     "",
		CouponName: "FLASH25",
	}

	mockService.On("ClaimCoupon", reqBody).Return(errors.New("user_id is required"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.ClaimCoupon(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "user_id is required", response["error"])

	mockService.AssertExpectations(t)
}

func TestGetCouponDetails_Handler_Success(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	expectedResponse := &models.CouponDetailResponse{
		Name:            "FLASH25",
		Amount:          100,
		RemainingAmount: 75,
		ClaimedBy:       []string{},
	}

	mockService.On("GetCouponDetails", "FLASH25").Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/coupons/FLASH25", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.GetCouponDetails)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.CouponDetailResponse
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "FLASH25", response.Name)
	assert.Equal(t, 100, response.Amount)
	assert.Equal(t, 75, response.RemainingAmount)

	mockService.AssertExpectations(t)
}

func TestGetCouponDetails_Handler_NotFound(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	mockService.On("GetCouponDetails", "NONEXISTENT").Return(nil, repository.ErrCouponNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/coupons/NONEXISTENT", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.GetCouponDetails)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestGetCouponDetails_Handler_InternalError(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	mockService.On("GetCouponDetails", "FLASH25").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/coupons/FLASH25", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.GetCouponDetails)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "database error", response["error"])

	mockService.AssertExpectations(t)
}

func TestUpdateCoupon_Handler_Success(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	mockService.On("UpdateCoupon", "FLASH25").Return(int64(1), nil)

	req := httptest.NewRequest(http.MethodPut, "/api/coupons/FLASH25", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.UpdateCoupon)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon updated successfully", response["message"])
	assert.Equal(t, float64(1), response["rows_affected"])

	mockService.AssertExpectations(t)
}

func TestUpdateCoupon_Handler_NotFound(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	mockService.On("UpdateCoupon", "NONEXISTENT").Return(int64(0), repository.ErrCouponNotFound)

	req := httptest.NewRequest(http.MethodPut, "/api/coupons/NONEXISTENT", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.UpdateCoupon)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Coupon not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestUpdateCoupon_Handler_InternalError(t *testing.T) {
	logger.Init()
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)

	mockService.On("UpdateCoupon", "FLASH25").Return(int64(0), errors.New("database error"))

	req := httptest.NewRequest(http.MethodPut, "/api/coupons/FLASH25", nil)
	rec := httptest.NewRecorder()

	// Use mux to inject path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/coupons/{name}", handler.UpdateCoupon)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "database error", response["error"])

	mockService.AssertExpectations(t)
}
