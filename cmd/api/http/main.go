package main

import (
	"context"
	"fmt"

	"github.com/wazadio/coupon-system/pkg/logger"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	if err := logger.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	StartServer(ctx, Init())

	logger.Log.Info("Server is shutting down...")
	logger.Log.Info("Goodbye!")
}
