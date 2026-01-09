package cmd

import (
	"github.com/wazadio/coupon-system/internal/database"
	"github.com/wazadio/coupon-system/internal/repository"
	"github.com/wazadio/coupon-system/internal/service"
)

type Deps struct {
	// Add dependencies here as needed

	// Repositories
	CouponRepository repository.CouponRepository

	// Services
	CouponService service.CouponService
}

func Init() (deps *Deps, err error) {
	deps = &Deps{}

	// Connect to the database
	db, err := database.Connect(database.NewConfigFromEnv())

	// Initialize repositories
	deps.CouponRepository = repository.NewCouponRepository(db)

	// Initialize services with injected repositories
	deps.CouponService = service.NewCouponService(deps.CouponRepository)

	return
}
