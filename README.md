# Flash Sale Coupon System

A scalable REST API built in Golang for managing flash sale coupons with high concurrency support and strict data consistency guarantees.

## Features

- **High Concurrency Support**: Handles multiple simultaneous requests safely using database transactions
- **Strict Data Consistency**: Guarantees no double-claiming and accurate stock management
- **Race Condition Prevention**: Uses PostgreSQL serializable transactions and row-level locking
- **RESTful API**: Clean API design following exact specifications
- **Docker Ready**: Complete Docker and Docker Compose setup for easy deployment

## Tech Stack

- **Language**: Go 1.21
- **Database**: PostgreSQL 15
- **Router**: Gorilla Mux v1.8.1 (HTTP routing)
- **Database Driver**: lib/pq v1.10.9 (PostgreSQL driver)
- **Logger**: Uber Zap v1.27.1 (structured logging)
- **Testing**: Testify v1.11.1, go-sqlmock v1.5.2
- **Infrastructure**: Docker & Docker Compose

## Prerequisites

- [Docker](https://www.docker.com/get-started) (version 20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (version 2.0+)

That's it! No need to install Go or PostgreSQL separately.

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/wazadio/coupon-system.git
cd coupon-system
```

### 2. Run the Application

```bash
docker-compose up --build
```

This command will:
- Build the Go application
- Start a PostgreSQL database
- Initialize the database schema
- Start the API server on port 8080

The application will be ready when you see:
```
coupon_api  | Server is ready to handle requests
```

### 3. Verify the Application is Running

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

## API Documentation

### 1. Create Coupon

Creates a new coupon in the system.

**Endpoint**: `POST /api/coupons`

**Request Body**:
```json
{
  "name": "PROMO_SUPER",
  "amount": 100
}
```

**Response**: `201 Created`

**Example**:
```bash
curl -X POST http://localhost:8080/api/coupons \
  -H "Content-Type: application/json" \
  -d '{"name":"PROMO_SUPER","amount":100}'
```

### 2. Claim Coupon

Attempts to claim a coupon for a specific user.

**Endpoint**: `POST /api/coupons/claim`

**Request Body**:
```json
{
  "user_id": "user_12345",
  "coupon_name": "PROMO_SUPER"
}
```

**Response Codes**:
- `200 OK`: Claim successful
- `409 Conflict`: User already claimed this coupon
- `400 Bad Request`: No stock available or invalid request
- `404 Not Found`: Coupon not found

**Example**:
```bash
curl -X POST http://localhost:8080/api/coupons/claim \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user_12345","coupon_name":"PROMO_SUPER"}'
```

### 3. Get Coupon Details

Retrieves coupon information including all users who claimed it.

**Endpoint**: `GET /api/coupons/{name}`

**Response**:
```json
{
  "name": "PROMO_SUPER",
  "amount": 100,
  "remaining_amount": 95,
  "claimed_by": ["user_12345", "user_67890"]
}
```

**Example**:
```bash
curl http://localhost:8080/api/coupons/PROMO_SUPER
```

### 4. Update Coupon

Updates a coupon's timestamp (utility endpoint).

**Endpoint**: `PUT /api/coupons/{name}` or `PATCH /api/coupons/{name}`

**Response**: `200 OK`
```json
{
  "message": "Coupon updated successfully",
  "rows_affected": 1
}
```

**Example**:
```bash
curl -X PUT http://localhost:8080/api/coupons/PROMO_SUPER
```

## Testing

### Unit Tests

The project has comprehensive unit tests for all layers:

```bash
# Run all unit tests
make test-unit

# Run tests with coverage summary
make test-coverage

# Or use go test directly
go test -v ./internal/...
```

Test coverage includes:
- **Handler Tests**: HTTP endpoint testing with mocked services (16 tests)
- **Model Tests**: Data structure validation (12 tests)
- **Repository Tests**: Database operations with SQL mocks (23+ tests)
- **Service Tests**: Business logic validation

### Integration Tests

Scenario-based tests that verify the system under high concurrency:

```bash
# Run integration tests (requires running server)
make test-scenarios

# Or run directly
go test -v ./test/...
```

Integration tests include:
- **Flash Sale Scenario**: 50 users competing for 5 items
- **Double Dip Scenario**: Same user attempting 10 concurrent claims

### Manual Testing

You can test the API using the provided examples above or use tools like Postman or HTTPie.

### Concurrency Testing

Test the "Flash Sale" scenario (50 concurrent requests for 5 items):

```bash
# Create a coupon with 5 items
curl -X POST http://localhost:8080/api/coupons \
  -H "Content-Type: application/json" \
  -d '{"name":"FLASH_SALE","amount":5}'

# Run 50 concurrent claim requests
for i in {1..50}; do
  curl -X POST http://localhost:8080/api/coupons/claim \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"user_$i\",\"coupon_name\":\"FLASH_SALE\"}" &
done
wait

# Check the results
curl http://localhost:8080/api/coupons/FLASH_SALE
```

Expected result: Exactly 5 claims, 0 remaining amount.

Test the "Double Dip" scenario (same user claiming 10 times):

```bash
# Create a coupon
curl -X POST http://localhost:8080/api/coupons \
  -H "Content-Type: application/json" \
  -d '{"name":"DOUBLE_DIP","amount":10}'

# Run 10 concurrent requests from the same user
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/coupons/claim \
    -H "Content-Type: application/json" \
    -d '{"user_id":"same_user","coupon_name":"DOUBLE_DIP"}' &
done
wait

# Check the results
curl http://localhost:8080/api/coupons/DOUBLE_DIP
```

Expected result: Exactly 1 claim from "same_user", 9 remaining.

## Architecture

### Database Design

The system uses two separate tables with proper constraints:

#### Coupons Table
```sql
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL,
    remaining_amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Claims Table
```sql
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, coupon_name),
    FOREIGN KEY (coupon_name) REFERENCES coupons(name) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX idx_claims_coupon_name ON claims(coupon_name);
CREATE INDEX idx_claims_user_id ON claims(user_id);
CREATE INDEX idx_claims_user_coupon ON claims(user_id, coupon_name);
```

**Key Design Decisions**:
- Separate tables for coupons and claims (no embedding)
- Composite unique constraint on `(user_id, coupon_name)` to prevent double-claiming
- Foreign key with CASCADE delete to maintain referential integrity
- Performance indexes for common query patterns

### Concurrency Strategy

The system handles high concurrency using:

1. **Serializable Transaction Isolation**: Prevents phantom reads and ensures strict consistency
2. **Row-Level Locking**: Uses `SELECT ... FOR UPDATE` to lock coupon rows during claims
3. **Atomic Operations**: All claim operations (check stock, insert claim, decrement stock) happen in a single transaction
4. **Unique Constraint**: Database-level enforcement of one claim per user per coupon

**Transaction Flow**:
```
BEGIN TRANSACTION (SERIALIZABLE)
  ↓
SELECT ... FOR UPDATE (Lock coupon row)
  ↓
Check remaining_amount > 0
  ↓
INSERT claim (fails if duplicate due to unique constraint)
  ↓
UPDATE remaining_amount - 1
  ↓
COMMIT TRANSACTION
```

This approach guarantees:
- No overselling (stock never goes negative)
- No double-claiming (enforced by unique constraint)
- Proper serialization of concurrent requests

### Project Structure

```
coupon-system/
├── cmd/
│   ├── init_resources.go          # Resource initialization
│   └── api/
│       └── http/
│           ├── main.go            # Application entry point
│           └── init.go            # Dependency injection setup
├── internal/
│   ├── database/
│   │   └── db.go                  # Database connection
│   ├── handlers/
│   │   ├── middleware/
│   │   │   └── logging.go         # HTTP logging middleware
│   │   └── rest/
│   │       ├── base_handler.go    # Base handler (health check)
│   │       ├── base_router.go     # Base routes
│   │       ├── coupon_handler.go  # Coupon HTTP handlers
│   │       ├── coupon_router.go   # Coupon routes
│   │       └── *_test.go          # Handler unit tests
│   ├── models/
│   │   ├── coupon.go              # Data models & DTOs
│   │   └── coupon_test.go         # Model tests
│   ├── repository/
│   │   ├── coupon_repository.go   # Database operations (interface)
│   │   └── coupon_repository_test.go  # Repository tests
│   └── service/
│       ├── coupon_service.go      # Business logic (interface)
│       └── coupon_service_test.go # Service tests
├── pkg/
│   ├── logger/
│   │   └── logger.go              # Structured logging (Zap)
│   └── rest/
│       └── rest.go                # HTTP response helpers
├── scripts/
│   ├── init.sql                   # Database schema
│   └── run_scenarios.go           # Scenario test runner
├── test/
│   ├── scenarios_test.go          # Integration tests
│   └── README.md                  # Test documentation
├── docs/
│   ├── ARCHITECTURE.md            # Architecture documentation
│   └── TESTING.md                 # Testing guide
├── docker-compose.yml
├── Dockerfile
├── Makefile                       # Build & test commands
├── go.mod
├── go.sum
└── README.md
```

### Architecture Highlights

- **Dependency Injection**: All components use interfaces for better testability
- **Layered Architecture**: Clear separation between handlers, services, and repositories
- **Middleware Support**: Request logging and potential for authentication/rate limiting
- **Comprehensive Testing**: Unit tests for all layers with >80% coverage

## Available Make Commands

The project includes a Makefile for common operations:

```bash
make help              # Show all available commands
make build             # Build Docker containers
make up                # Start application with build
make up-detached       # Start in background
make down              # Stop and remove containers
make restart           # Restart the application
make logs              # View all logs
make logs-api          # View API logs only
make logs-db           # View database logs only
make test-unit         # Run unit tests with coverage
make test-coverage     # Run tests with coverage summary
make test-scenarios    # Run integration tests
make clean             # Clean up Docker resources
```

## Stopping the Application

Press `Ctrl+C` in the terminal where docker-compose is running, or run:

```bash
make down
# or
docker-compose down
```

To remove all data (including database volumes):

```bash
docker-compose down -v
```

## Environment Variables

The following environment variables can be configured in `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | postgres | Database host |
| DB_PORT | 5432 | Database port |
| DB_USER | coupon_user | Database username |
| DB_PASSWORD | coupon_pass | Database password |
| DB_NAME | coupon_db | Database name |
| SERVER_PORT | 8080 | API server port |

## Troubleshooting

### Database Connection Issues

If you see database connection errors, wait a few seconds for PostgreSQL to fully initialize. The API will retry the connection automatically.

### Port Already in Use

If port 8080 or 5432 is already in use, modify the ports in `docker-compose.yml`:

```yaml
services:
  api:
    ports:
      - "8081:8080"  # Change host port
```

### View Logs

```bash
# All services
docker-compose logs

# API only
docker-compose logs api

# Database only
docker-compose logs postgres

# Follow logs in real-time
docker-compose logs -f
```

## Development

To run the application locally without Docker:

1. Install Go 1.21+
2. Install PostgreSQL 15+
3. Set environment variables
4. Run migrations: `psql -U coupon_user -d coupon_db -f scripts/init.sql`
5. Run the application: `go run cmd/api/main.go`

## Additional Documentation

- [Architecture Details](docs/ARCHITECTURE.md) - In-depth architecture and design patterns
- [Testing Guide](docs/TESTING.md) - Comprehensive testing documentation
- [Integration Tests](test/README.md) - Scenario-based test documentation

## Key Features Implemented

✅ RESTful API with 4 endpoints (Create, Claim, Get, Update)
✅ High concurrency handling with serializable transactions
✅ Race condition prevention with row-level locking
✅ Comprehensive unit tests (50+ tests, >80% coverage)
✅ Integration tests for critical scenarios
✅ Structured logging with Uber Zap
✅ Dependency injection for testability
✅ Docker containerization
✅ Database migrations and indexing
✅ Request logging middleware
✅ Error handling and validation
