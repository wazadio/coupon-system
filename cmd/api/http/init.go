package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/wazadio/coupon-system/cmd"
	"github.com/wazadio/coupon-system/internal/handlers/middleware"
	"github.com/wazadio/coupon-system/internal/handlers/rest"
	"github.com/wazadio/coupon-system/pkg/logger"
	"go.uber.org/zap"
)

type handler interface {
	SetupRouter(*mux.Router)
}

func Init() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	deps, err := cmd.Init()
	if err != nil {
		panic(err)
	}

	// API subrouter with /api prefix
	api := router.PathPrefix("/api").Subrouter()

	// Initialize and setup routers for different handlers
	var handlers []handler

	handlers = append(handlers, rest.NewCouponHandler(deps.CouponService))
	handlers = append(handlers, &rest.BaseHandler{})

	for _, handler := range handlers {
		handler.SetupRouter(api)
	}

	router.Use(middleware.LoggingMiddleware)

	return router
}

func StartServer(ctx context.Context, router *mux.Router) error {
	port := "8080"
	if p := os.Getenv("SERVER_PORT"); p != "" {
		port = p
	}

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Log.Error("Server error", zap.Error(err))
		}
	}()

	logger.Log.Info("Server is listening", zap.String("port", port))

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
