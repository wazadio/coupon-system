# Coupon System Architecture

## Dependency Injection Pattern

This application has been refactored to use the **Dependency Injection (DI)** pattern throughout the codebase. This provides better testability, maintainability, and flexibility.

## Architecture Overview

```
┌─────────────────┐
│   main.go       │  - Application entry point
│                 │  - Wires all dependencies
└────────┬────────┘
         │
         ├──────────────────────────────────────┐
         │                                      │
         v                                      v
┌─────────────────┐                  ┌─────────────────┐
│   Database      │                  │   Handlers      │
│   (sql.DB)      │                  │  (HTTP Layer)   │
└────────┬────────┘                  └────────┬────────┘
         │                                      │
         │ inject                               │ inject
         v                                      v
┌─────────────────┐                  ┌─────────────────┐
│  Repository     │  <───inject───── │   Service       │
│  (Interface)    │                  │  (Interface)    │
└─────────────────┘                  └─────────────────┘
```

## Key Components

### 1. Database Layer (`internal/database/`)

**Before (Global State):**
```go
var DB *sql.DB

func InitDB() error {
    DB, err = sql.Open(...)
}
```

**After (Dependency Injection):**
```go
type Config struct {
    Host, Port, User, Password, DBName string
}

func Connect(config *Config) (*sql.DB, error) {
    return sql.Open(...)
}
```

**Benefits:**
- No global state
- Easier to test with mock databases
- Configuration is explicit and testable

### 2. Repository Layer (`internal/repository/`)

**Before:**
```go
type CouponRepository struct {
    db *sql.DB
}

func NewCouponRepository() *CouponRepository {
    return &CouponRepository{
        db: database.DB,  // Global dependency
    }
}
```

**After:**
```go
// Interface for abstraction
type CouponRepository interface {
    CreateCoupon(name string, amount int) error
    ClaimCoupon(userID, couponName string) error
    GetCouponByName(name string) (*models.CouponDetailResponse, error)
}

// Implementation with injected dependency
type couponRepository struct {
    db *sql.DB
}

func NewCouponRepository(db *sql.DB) CouponRepository {
    return &couponRepository{
        db: db,  // Injected dependency
    }
}
```

**Benefits:**
- Interface allows for easy mocking in tests
- Database connection is explicitly injected
- Can swap implementations without changing consumers
- Private struct enforces interface usage

### 3. Service Layer (`internal/service/`)

**Before:**
```go
type CouponService struct {
    repo *repository.CouponRepository
}

func NewCouponService() *CouponService {
    return &CouponService{
        repo: repository.NewCouponRepository(),  // Creates its own dependency
    }
}
```

**After:**
```go
// Interface for abstraction
type CouponService interface {
    CreateCoupon(req *models.CreateCouponRequest) error
    ClaimCoupon(req *models.ClaimCouponRequest) error
    GetCouponDetails(name string) (*models.CouponDetailResponse, error)
}

// Implementation with injected dependency
type couponService struct {
    repo repository.CouponRepository
}

func NewCouponService(repo repository.CouponRepository) CouponService {
    return &couponService{
        repo: repo,  // Injected dependency
    }
}
```

**Benefits:**
- Interface allows for easy mocking in tests
- Repository is explicitly injected
- Can test service logic independently
- Private struct enforces interface usage

### 4. Handler Layer (`internal/handlers/`)

**Before:**
```go
type CouponHandler struct {
    service *service.CouponService
}

func NewCouponHandler() *CouponHandler {
    return &CouponHandler{
        service: service.NewCouponService(),  // Creates its own dependency
    }
}

func SetupRouter() *mux.Router {
    handler := NewCouponHandler()
    // ...
}
```

**After:**
```go
type CouponHandler struct {
    service service.CouponService
}

func NewCouponHandler(service service.CouponService) *CouponHandler {
    return &CouponHandler{
        service: service,  // Injected dependency
    }
}

func SetupRouter(couponService service.CouponService) *mux.Router {
    handler := NewCouponHandler(couponService)
    // ...
}
```

**Benefits:**
- Service is explicitly injected
- Handlers can be tested with mock services
- Clear dependency flow

### 5. Main Application (`cmd/api/main.go`)

**Before:**
```go
func main() {
    database.InitDB()
    defer database.CloseDB()
    
    router := handlers.SetupRouter()
    // ...
}
```

**After:**
```go
func main() {
    // 1. Initialize database connection
    dbConfig := database.NewConfigFromEnv()
    db, err := database.Connect(dbConfig)
    defer db.Close()
    
    // 2. Build dependency chain (bottom-up)
    couponRepo := repository.NewCouponRepository(db)
    couponService := service.NewCouponService(couponRepo)
    
    // 3. Setup router with injected dependencies
    router := handlers.SetupRouter(couponService)
    // ...
}
```

**Benefits:**
- Clear dependency wiring in one place
- Easy to see the entire dependency graph
- Simple to add new dependencies
- Each component receives exactly what it needs

## Benefits of This Architecture

### 1. **Testability**
- Each layer can be tested independently with mocks
- No need to setup real database for unit tests
- Example test structure:
  ```go
  type MockRepository struct{}
  func (m *MockRepository) CreateCoupon(name string, amount int) error {
      return nil
  }
  
  func TestCouponService(t *testing.T) {
      mockRepo := &MockRepository{}
      service := NewCouponService(mockRepo)
      // Test service logic without database
  }
  ```

### 2. **Flexibility**
- Easy to swap implementations (e.g., PostgreSQL → MongoDB)
- Can add caching layer without changing existing code
- Multiple implementations of same interface

### 3. **Maintainability**
- Clear separation of concerns
- Dependencies are explicit and visible
- Changes to one layer don't cascade to others
- Easier to understand code flow

### 4. **No Global State**
- No global variables (`database.DB` removed)
- Thread-safe by design
- Easier to reason about code

### 5. **SOLID Principles**
- **Single Responsibility**: Each layer has one job
- **Open/Closed**: Open for extension, closed for modification
- **Liskov Substitution**: Interfaces can be substituted
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions (interfaces)

## Dependency Flow

```
main.go
  ↓ creates
Database Connection (*sql.DB)
  ↓ injected into
Repository (CouponRepository interface)
  ↓ injected into
Service (CouponService interface)
  ↓ injected into
Handler (CouponHandler)
  ↓ used by
Router (HTTP Server)
```

## Testing Strategy

### Unit Tests
- **Repository**: Use `sqlmock` or test database
- **Service**: Use mock repository
- **Handler**: Use mock service

### Integration Tests
- Wire up real components
- Use test database
- Test complete flows

### Example Mock
```go
// Mock repository for testing service
type mockCouponRepository struct{}

func (m *mockCouponRepository) CreateCoupon(name string, amount int) error {
    return nil
}

func (m *mockCouponRepository) ClaimCoupon(userID, couponName string) error {
    return repository.ErrAlreadyClaimed
}

func (m *mockCouponRepository) GetCouponByName(name string) (*models.CouponDetailResponse, error) {
    return &models.CouponDetailResponse{
        Name:            name,
        Amount:          10,
        RemainingAmount: 5,
        ClaimedBy:       []string{"user1", "user2"},
    }, nil
}
```

## Future Enhancements

With DI in place, it's easy to add:

1. **Caching Layer**
   ```go
   type CachedCouponRepository struct {
       repo  CouponRepository
       cache Cache
   }
   ```

2. **Logging/Metrics**
   ```go
   type LoggedCouponService struct {
       service CouponService
       logger  Logger
   }
   ```

3. **Multiple Database Support**
   ```go
   // PostgresRepository
   // MongoDBRepository
   // Both implement CouponRepository interface
   ```

4. **Feature Flags**
   ```go
   func main() {
       if os.Getenv("USE_CACHE") == "true" {
           repo = NewCachedRepository(db, cache)
       } else {
           repo = NewCouponRepository(db)
       }
   }
   ```

## Conclusion

The Dependency Injection pattern makes the codebase:
- ✅ More testable
- ✅ More maintainable
- ✅ More flexible
- ✅ Follows SOLID principles
- ✅ Production-ready for scaling

All dependencies are now explicit, making the code easier to understand, test, and extend.
