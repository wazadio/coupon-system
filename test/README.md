# Coupon System - Critical Scenario Tests

This directory contains automated tests for the two critical scenarios required by the technical assessment.

## Test Scenarios

### 1. Flash Sale Attack (`TestFlashSaleScenario`)
Tests high concurrency with limited stock:
- Creates a coupon with only **5 items** in stock
- Launches **50 concurrent requests** to claim the coupon
- **Expected Result**: Exactly 5 successful claims, 45 failures, 0 remaining stock

### 2. Double Dip Attack (`TestDoubleDipScenario`)
Tests preventing duplicate claims from the same user:
- Creates a coupon with plenty of stock (100 items)
- Launches **10 concurrent requests** from the **SAME user**
- **Expected Result**: Exactly 1 successful claim, 9 failures (409 Conflict)

## Prerequisites

1. The coupon system must be running:
```bash
docker-compose up --build
```

2. Wait for the system to be fully ready (database initialized, API responding)

## Running the Tests

### Option 1: Using Go Test Command
```bash
# Run all tests
go test -v ./test/...

# Run a specific test
go test -v ./test -run TestFlashSaleScenario
go test -v ./test -run TestDoubleDipScenario

# Run with timeout
go test -v -timeout 30s ./test/...
```

### Option 2: Using Make (if Makefile exists)
```bash
make test
```

### Option 3: Run from test directory
```bash
cd test
go test -v
```

## Test Output

The tests provide detailed output including:
- Setup confirmation
- Real-time progress of concurrent requests
- Status code distribution
- Success/failure counts
- Coupon details verification
- Final PASS/FAIL verdict

Example output:
```
=== Testing Flash Sale Attack Scenario ===
Setup: Creating coupon 'FLASH_SALE_TEST' with 5 items
✓ Coupon created successfully
Launching 50 concurrent claim requests...
✓ User user_0 successfully claimed coupon (status: 200)
✗ User user_1 failed to claim coupon (status: 400)
...
All requests completed in 245ms
Status code distribution: map[200:5 400:45]
Successful claims: 5 (expected: 5)
Failed claims: 45 (expected: 45)
Coupon details: Amount=5, Remaining=0, Claimed by 5 users
✓✓✓ Flash Sale Attack Test PASSED ✓✓✓
```

## What the Tests Verify

### Flash Sale Test Verifies:
- ✓ Exactly 5 successful claims (no over-claiming)
- ✓ Exactly 45 failed claims
- ✓ Remaining stock is 0
- ✓ `claimed_by` array contains exactly 5 users
- ✓ Atomic transaction handling prevents race conditions

### Double Dip Test Verifies:
- ✓ Exactly 1 successful claim from the same user
- ✓ Exactly 9 rejections with appropriate status codes (409/400)
- ✓ User appears only once in `claimed_by` array
- ✓ Database uniqueness constraint on (user_id, coupon_name) works
- ✓ Remaining stock decreased by only 1

## Troubleshooting

### Connection Refused Error
```
Failed to send request: connection refused
```
**Solution**: Ensure the API is running on http://localhost:8080

### Tests Failing
1. Check if the database is properly initialized
2. Verify the API endpoints are working:
   ```bash
   curl http://localhost:8080/health
   ```
3. Check Docker logs:
   ```bash
   docker-compose logs api
   docker-compose logs postgres
   ```

### Inconsistent Results
If tests pass sometimes but fail other times, this indicates race conditions in the implementation. The atomic transaction handling needs to be fixed.

## Configuration

You can modify the test parameters in `scenarios_test.go`:
```go
const (
    baseURL           = "http://localhost:8080/api"  // API base URL
    flashSaleStock    = 5                            // Stock for flash sale test
    concurrentFlash   = 50                           // Concurrent requests for flash sale
    concurrentDoubleDip = 10                         // Concurrent requests for double dip
)
```
