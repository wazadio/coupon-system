package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	baseURL = "http://localhost:8080/api"
)

// CouponRequest represents the request to create a coupon
type CouponRequest struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

// ClaimRequest represents the request to claim a coupon
type ClaimRequest struct {
	UserID     string `json:"user_id"`
	CouponName string `json:"coupon_name"`
}

// CouponDetails represents the coupon details response
type CouponDetails struct {
	Name            string   `json:"name"`
	Amount          int      `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}

// TestFlashSaleScenario tests the Flash Sale Attack scenario
// 50 users try to claim a coupon with only 5 items
// Expected: Exactly 5 succeed, 45 fail, no overselling
func TestFlashSaleScenario(t *testing.T) {
	if !isServerReady(t) {
		t.Skip("Server not ready, skipping integration test")
	}

	couponName := "FLASH_SALE_TEST"
	stock := 5
	concurrentUsers := 50

	t.Log("=== Testing Flash Sale Attack Scenario ===")
	t.Logf("Setup: Creating coupon '%s' with %d items", couponName, stock)

	// Create coupon
	err := createCoupon(couponName, stock)
	if err != nil {
		t.Fatalf("Failed to create coupon: %v", err)
	}

	t.Logf("Launching %d concurrent claim requests...", concurrentUsers)

	// Launch concurrent claims
	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex
	statusCodes := make(map[int]int)

	startTime := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			userID := fmt.Sprintf("user_%d", userNum)
			statusCode, _ := claimCoupon(userID, couponName)

			mu.Lock()
			defer mu.Unlock()

			statusCodes[statusCode]++

			if statusCode == 200 {
				successCount++
			} else {
				failureCount++
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("All requests completed in %v", duration)
	t.Logf("Status code distribution: %v", statusCodes)

	// Verify results
	details, err := getCouponDetails(couponName)
	assert.NoError(t, err, "Failed to get coupon details")
	assert.NotNil(t, details, "Coupon details should not be nil")

	t.Logf("Results:")
	t.Logf("  Successful claims: %d (expected: %d)", successCount, stock)
	t.Logf("  Failed claims: %d (expected: %d)", failureCount, concurrentUsers-stock)
	t.Logf("  Remaining amount: %d (expected: 0)", details.RemainingAmount)
	t.Logf("  Unique claimers: %d (expected: %d)", len(details.ClaimedBy), stock)

	// Assertions
	assert.Equal(t, stock, successCount, "Exactly %d claims should succeed", stock)
	assert.Equal(t, concurrentUsers-stock, failureCount, "Exactly %d claims should fail", concurrentUsers-stock)
	assert.Equal(t, 0, details.RemainingAmount, "Remaining amount should be 0")
	assert.Equal(t, stock, len(details.ClaimedBy), "Should have exactly %d unique claimers", stock)

	// Verify no duplicates in claimed_by
	uniqueUsers := make(map[string]bool)
	for _, userID := range details.ClaimedBy {
		assert.False(t, uniqueUsers[userID], "User %s should appear only once in claimed_by", userID)
		uniqueUsers[userID] = true
	}

	t.Log("✅ Flash Sale scenario PASSED - No overselling detected")
}

// TestDoubleDipScenario tests the Double Dip Attack scenario
// Same user tries to claim the same coupon multiple times concurrently
// Expected: Only 1 claim succeeds, all others fail with 409 Conflict
func TestDoubleDipScenario(t *testing.T) {
	if !isServerReady(t) {
		t.Skip("Server not ready, skipping integration test")
	}

	couponName := "DOUBLE_DIP_TEST"
	stock := 100
	concurrentAttempts := 10
	sameUserID := "same_user_123"

	t.Log("=== Testing Double Dip Attack Scenario ===")
	t.Logf("Setup: Creating coupon '%s' with enough stock", couponName)

	// Create coupon
	err := createCoupon(couponName, stock)
	if err != nil {
		t.Fatalf("Failed to create coupon: %v", err)
	}

	t.Logf("Launching %d concurrent claims from user '%s'...", concurrentAttempts, sameUserID)

	// Launch concurrent claims from SAME user
	var wg sync.WaitGroup
	successCount := 0
	conflictCount := 0
	var mu sync.Mutex
	statusCodes := make(map[int]int)

	startTime := time.Now()

	for i := 0; i < concurrentAttempts; i++ {
		wg.Add(1)
		go func(attemptNum int) {
			defer wg.Done()

			statusCode, _ := claimCoupon(sameUserID, couponName)

			mu.Lock()
			defer mu.Unlock()

			statusCodes[statusCode]++

			if statusCode == 200 {
				successCount++
			} else if statusCode == 409 {
				conflictCount++
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("All requests completed in %v", duration)
	t.Logf("Status code distribution: %v", statusCodes)

	// Verify results
	details, err := getCouponDetails(couponName)
	assert.NoError(t, err, "Failed to get coupon details")
	assert.NotNil(t, details, "Coupon details should not be nil")

	// Count user appearances
	userAppearances := 0
	for _, claimedUserID := range details.ClaimedBy {
		if claimedUserID == sameUserID {
			userAppearances++
		}
	}

	t.Logf("Results:")
	t.Logf("  Successful claims: %d (expected: 1)", successCount)
	t.Logf("  Conflict responses: %d (expected: %d)", conflictCount, concurrentAttempts-1)
	t.Logf("  User appearances in claimed_by: %d (expected: 1)", userAppearances)
	t.Logf("  Remaining amount: %d (expected: %d)", details.RemainingAmount, stock-1)

	// Assertions
	assert.Equal(t, 1, successCount, "Exactly 1 claim should succeed")
	assert.Equal(t, 1, userAppearances, "User should appear exactly once in claimed_by")
	assert.Equal(t, stock-1, details.RemainingAmount, "Remaining amount should be %d", stock-1)

	t.Log("✅ Double Dip scenario PASSED - Duplicate claims prevented")
}

// Helper function to check if server is ready
func isServerReady(t *testing.T) bool {
	resp, err := http.Get(baseURL + "/../health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// Helper function to create a coupon
func createCoupon(name string, amount int) error {
	reqBody := CouponRequest{
		Name:   name,
		Amount: amount,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(
		baseURL+"/coupons",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Helper function to claim a coupon
func claimCoupon(userID, couponName string) (int, error) {
	reqBody := ClaimRequest{
		UserID:     userID,
		CouponName: couponName,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(
		baseURL+"/coupons/claim",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// Helper function to get coupon details
func getCouponDetails(name string) (*CouponDetails, error) {
	resp, err := http.Get(baseURL + "/coupons/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var details CouponDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &details, nil
}
