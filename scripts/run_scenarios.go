package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	baseURL             = "http://localhost:8080/api"
	flashSaleCoupon     = "FLASH_SALE_TEST"
	doubleDipCoupon     = "DOUBLE_DIP_TEST"
	flashSaleStock      = 5
	concurrentFlash     = 50
	concurrentDoubleDip = 10
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

// ScenarioResult holds the result of a test scenario
type ScenarioResult struct {
	Name     string
	Success  bool
	Duration time.Duration
	Message  string
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     Coupon System - Concurrent Test Scenarios Runner      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Check if server is ready
	if !waitForServer() {
		fmt.Println("❌ Server is not responding. Please start the server first.")
		os.Exit(1)
	}

	// Run scenarios concurrently
	var wg sync.WaitGroup
	results := make(chan ScenarioResult, 2)

	// Launch Flash Sale scenario
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := runFlashSaleScenario()
		results <- result
	}()

	// Launch Double Dip scenario
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := runDoubleDipScenario()
		results <- result
	}()

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and display results
	var scenarioResults []ScenarioResult
	for result := range results {
		scenarioResults = append(scenarioResults, result)
	}

	// Print summary
	printSummary(scenarioResults)
}

// waitForServer checks if the server is ready
func waitForServer() bool {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("✅ Server is ready")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Printf("⏳ Waiting for server... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}
	return false
}

// runFlashSaleScenario runs the Flash Sale Attack scenario
func runFlashSaleScenario() ScenarioResult {
	result := ScenarioResult{
		Name:    "Flash Sale Attack",
		Success: false,
	}
	startTime := time.Now()

	fmt.Println("\n┌────────────────────────────────────────────────────────────┐")
	fmt.Println("│          Running: Flash Sale Attack Scenario              │")
	fmt.Println("└────────────────────────────────────────────────────────────┘")
	fmt.Printf("📦 Creating coupon '%s' with %d items\n", flashSaleCoupon, flashSaleStock)

	// Step 1: Create the coupon
	if err := createCoupon(flashSaleCoupon, flashSaleStock); err != nil {
		result.Message = fmt.Sprintf("Failed to create coupon: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	fmt.Println("✅ Coupon created successfully")

	// Step 2: Launch concurrent claims
	fmt.Printf("🚀 Launching %d concurrent claim requests...\n", concurrentFlash)

	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex
	statusCodes := make(map[int]int)

	claimStart := time.Now()

	for i := 0; i < concurrentFlash; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			userID := fmt.Sprintf("user_%d", userNum)
			statusCode, _ := claimCoupon(userID, flashSaleCoupon)

			mu.Lock()
			defer mu.Unlock()

			statusCodes[statusCode]++

			if statusCode == 200 || statusCode == 201 {
				successCount++
			} else {
				failureCount++
			}
		}(i)
	}

	wg.Wait()
	claimDuration := time.Since(claimStart)

	fmt.Printf("⏱️  All requests completed in %v\n", claimDuration)
	fmt.Printf("📊 Status code distribution: %v\n", statusCodes)

	// Step 3: Verify results
	fmt.Println("🔍 Verifying results...")
	details, err := getCouponDetails(flashSaleCoupon)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to get coupon details: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	fmt.Printf("   ✓ Successful claims: %d (expected: %d)\n", successCount, flashSaleStock)
	fmt.Printf("   ✓ Failed claims: %d (expected: %d)\n", failureCount, concurrentFlash-flashSaleStock)
	fmt.Printf("   ✓ Remaining amount: %d (expected: 0)\n", details.RemainingAmount)
	fmt.Printf("   ✓ Unique claimers: %d (expected: %d)\n", len(details.ClaimedBy), flashSaleStock)

	// Validate
	result.Success = successCount == flashSaleStock &&
		failureCount == (concurrentFlash-flashSaleStock) &&
		details.RemainingAmount == 0 &&
		len(details.ClaimedBy) == flashSaleStock

	if result.Success {
		result.Message = "✅ PASSED - Exactly 5 claims succeeded, no overselling"
	} else {
		result.Message = fmt.Sprintf("❌ FAILED - Success: %d, Failures: %d, Remaining: %d",
			successCount, failureCount, details.RemainingAmount)
	}

	result.Duration = time.Since(startTime)
	return result
}

// runDoubleDipScenario runs the Double Dip Attack scenario
func runDoubleDipScenario() ScenarioResult {
	result := ScenarioResult{
		Name:    "Double Dip Attack",
		Success: false,
	}
	startTime := time.Now()

	fmt.Println("\n┌────────────────────────────────────────────────────────────┐")
	fmt.Println("│          Running: Double Dip Attack Scenario               │")
	fmt.Println("└────────────────────────────────────────────────────────────┘")
	fmt.Printf("📦 Creating coupon '%s' with enough stock\n", doubleDipCoupon)

	// Step 1: Create the coupon
	if err := createCoupon(doubleDipCoupon, 100); err != nil {
		result.Message = fmt.Sprintf("Failed to create coupon: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	fmt.Println("✅ Coupon created successfully")

	// Step 2: Launch concurrent claims from SAME user
	sameUserID := "same_user_123"
	fmt.Printf("🚀 Launching %d concurrent claims from user '%s'...\n", concurrentDoubleDip, sameUserID)

	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex
	statusCodes := make(map[int]int)

	claimStart := time.Now()

	for i := 0; i < concurrentDoubleDip; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()

			statusCode, _ := claimCoupon(sameUserID, doubleDipCoupon)

			mu.Lock()
			defer mu.Unlock()

			statusCodes[statusCode]++

			if statusCode == 200 || statusCode == 201 {
				successCount++
			} else {
				failureCount++
			}
		}(i)
	}

	wg.Wait()
	claimDuration := time.Since(claimStart)

	fmt.Printf("⏱️  All requests completed in %v\n", claimDuration)
	fmt.Printf("📊 Status code distribution: %v\n", statusCodes)

	// Step 3: Verify results
	fmt.Println("🔍 Verifying results...")
	details, err := getCouponDetails(doubleDipCoupon)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to get coupon details: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Count user appearances
	userAppearances := 0
	for _, claimedUserID := range details.ClaimedBy {
		if claimedUserID == sameUserID {
			userAppearances++
		}
	}

	fmt.Printf("   ✓ Successful claims: %d (expected: 1)\n", successCount)
	fmt.Printf("   ✓ Failed claims: %d (expected: %d)\n", failureCount, concurrentDoubleDip-1)
	fmt.Printf("   ✓ User appearances in claimed_by: %d (expected: 1)\n", userAppearances)
	fmt.Printf("   ✓ Remaining amount: %d (expected: 99)\n", details.RemainingAmount)

	// Validate
	result.Success = successCount == 1 &&
		failureCount == (concurrentDoubleDip-1) &&
		userAppearances == 1 &&
		details.RemainingAmount == 99

	if result.Success {
		result.Message = "✅ PASSED - Only 1 claim succeeded, duplicate prevented"
	} else {
		result.Message = fmt.Sprintf("❌ FAILED - Success: %d, Failures: %d, User appearances: %d",
			successCount, failureCount, userAppearances)
	}

	result.Duration = time.Since(startTime)
	return result
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

// printSummary prints a summary of all test results
func printSummary(results []ScenarioResult) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    TEST SUMMARY                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	passedCount := 0
	failedCount := 0

	for _, result := range results {
		fmt.Println()
		fmt.Printf("Test: %s\n", result.Name)
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Printf("Result: %s\n", result.Message)

		if result.Success {
			passedCount++
		} else {
			failedCount++
		}
	}

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n",
		len(results), passedCount, failedCount)
	fmt.Println(strings.Repeat("─", 60))

	if failedCount == 0 {
		fmt.Println("\n🎉 ALL TESTS PASSED! 🎉")
		os.Exit(0)
	} else {
		fmt.Println("\n⚠️  SOME TESTS FAILED")
		os.Exit(1)
	}
}
