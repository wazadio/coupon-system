package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/wazadio/coupon-system/pkg/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware adds trace_id to each request and creates a contextual logger
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate trace ID
		traceID := uuid.New().String()

		// Create a child logger with trace_id
		reqLogger := logger.Log.With(
			zap.String("trace_id", traceID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		// Add logger and trace_id to context
		ctx := context.WithValue(r.Context(), logger.LoggerContext{}, reqLogger)

		// Add trace_id to response header
		w.Header().Set("X-Trace-ID", traceID)

		// Log request
		reqLogger.Info("Request started")

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))

		// Log request completed
		reqLogger.Info("Request completed")
	})
}
