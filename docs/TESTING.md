# Unit Testing Strategy - TDD with 100% Coverage

## Overview
This document outlines the unit testing approach for the coupon system using Test-Driven Development (TDD) principles.

## Test Structure

### Required Test Files (To be created):
```
internal/
├── repository/
│   └── coupon_repository_test.go       ✅ Created (needs deps)
├── service/
│   └── coupon_service_test.go          ✅ Created (needs fixes)
├── handlers/
│   └── rest/
│       ├── coupon_handler_test.go      ✅ Created (needs fixes)
│       └── base_handler_test.go        ✅ Created
├── models/
│   └── coupon_test.go                  ✅ Created (needs fixes)
├── database/
│   └── db_test.go                      ❌ To create
└── handlers/
    └── middleware/
        └── logging_test.go             ❌ To create
pkg/
└── logger/
    └── logger_test.go                  ❌ To create
```

## Dependencies Required

```bash
go get github.com/DATA-DOG/go-sqlmock
go get github.com/stretchr/testify
```

## Testing Layers

### 1. Repository Layer Tests (`coupon_repository_test.go`)
**Coverage Target: 100%**

Tests:
- `TestCreateCoupon`
  - ✅ Success case
  - ✅ Duplicate coupon (23505 constraint violation)
  - ✅ Database error
  
- `TestClaimCoupon`
  - ✅ Success with transaction
  - ✅ Coupon not found
  - ✅ No stock available
  - ✅ Already claimed (unique constraint)
  - ✅ Transaction begin error
  - ✅ SELECT error
  - ✅ INSERT claim error
  - ✅ UPDATE coupon error
  - ✅ Commit error
  
- `TestGetCouponByName`
  - ✅ Success with claims
  - ✅ Success without claims
  - ✅ Coupon not found
  - ✅ Query error
  - ✅ Claims query error
  - ✅ Scan error

**Tools**: `go-sqlmock` for database mocking

### 2. Service Layer Tests (`coupon_service_test.go`)
**Coverage Target: 100%**

Tests:
- `TestCreateCoupon`
  - ✅ Success
  - ✅ Invalid amount (0, negative)
  - ✅ Empty name
  - ✅ Already exists error
  - ✅ Repository error

- `TestClaimCoupon`
  - ✅ Success
  - ✅ Empty user ID
  - ✅ Empty coupon name
  - ✅ Coupon not found
  - ✅ Already claimed
  - ✅ No stock
  - ✅ Repository error

- `TestGetCouponDetails`
  - ✅ Success
  - ✅ Empty name
  - ✅ Not found
  - ✅ Repository error

**Tools**: `testify/mock` for service mocking

### 3. Handler Layer Tests (`coupon_handler_test.go`, `base_handler_test.go`)
**Coverage Target: 100%**

Tests for each endpoint:
- `TestCreateCoupon_Handler`
  - ✅ Success (201)
  - ✅ Invalid JSON (400)
  - ✅ Validation error (400)
  - ✅ Already exists (409)
  - ✅ Internal error (500)

- `TestClaimCoupon_Handler`
  - ✅ Success (200)
  - ✅ Invalid JSON (400)
  - ✅ Validation error (400)
  - ✅ Not found (404)
  - ✅ Already claimed (409)
  - ✅ No stock (400)

- `TestGetCouponDetails_Handler`
  - ✅ Success (200)
  - ✅ Not found (404)
  - ✅ Internal error (500)

- `TestHealthCheck`
  - ✅ Returns healthy status

**Tools**: `httptest` for HTTP testing, `testify/mock` for service mocking

### 4. Model Tests (`coupon_test.go`)
**Coverage Target: 100%**

Tests:
- Struct initialization
- Field validation
- Edge cases (nil, empty)

### 5. Middleware Tests (`logging_test.go`)
**Coverage Target: 100%**

Tests:
- Request logging with trace_id
- Context propagation
- Error scenarios

### 6. Logger Tests (`logger_test.go`)
**Coverage Target: 100%**

Tests:
- Logger initialization
- File writing
- Log levels
- Context logging

### 7. Database Tests (`db_test.go`)
**Coverage Target: 100%**

Tests:
- Connection success
- Connection failure
- Config validation

## Running Tests

```bash
# Run all unit tests
make test-unit

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/repository/...

# Run with coverage report
go test -cover -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go test -cover ./internal/... | grep coverage
```

## Coverage Goals

| Package | Target | Status |
|---------|--------|--------|
| repository | 100% | ⏳ In Progress |
| service | 100% | ⏳ In Progress |
| handlers | 100% | ⏳ In Progress |
| models | 100% | ⏳ In Progress |
| database | 100% | ❌ Not Started |
| middleware | 100% | ❌ Not Started |
| logger | 100% | ❌ Not Started |

## TDD Principles Applied

1. **Red-Green-Refactor**
   - Write failing test first
   - Write minimal code to pass
   - Refactor while keeping tests green

2. **Test Coverage**
   - All public functions tested
   - All error paths tested
   - Edge cases covered

3. **Test Isolation**
   - Use mocks for dependencies
   - No external dependencies in unit tests
   - Each test is independent

4. **Clear Test Names**
   - Format: `Test<FunctionName>_<Scenario>`
   - Example: `TestCreateCoupon_Success`

## Next Steps

1. Fix test file syntax errors
2. Create missing test files
3. Run coverage analysis
4. Achieve 100% coverage
5. Add integration tests

## Commands Cheat Sheet

```bash
# Install dependencies
go get github.com/DATA-DOG/go-sqlmock
go get github.com/stretchr/testify

# Run tests
go test ./...

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestCreateCoupon ./internal/repository

# Verbose output
go test -v ./...

# Show only failing tests
go test -v ./... | grep FAIL
```
