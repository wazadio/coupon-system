package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	handler := &BaseHandler{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.healthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"healthy"}`, w.Body.String())
}

func TestBaseHandler_SetupRouter(t *testing.T) {
	handler := &BaseHandler{}
	router := mux.NewRouter()

	handler.SetupRouter(router)

	// Test that health route is registered
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"healthy"}`, w.Body.String())
}
