package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Mock implementation of the ServicoBUseCase interface
type MockWeatherUseCase struct {
	ExecuteFn      func(zipcode string) (map[string]float64, error)
	ExecuteCalled  bool
	ExecuteZipcode string
}

func (m *MockWeatherUseCase) Execute(zipcode string) (map[string]float64, error) {
	m.ExecuteCalled = true
	m.ExecuteZipcode = zipcode
	return m.ExecuteFn(zipcode)
}

func TestWeatherHandler_GetWeather_Success(t *testing.T) {
	// Setup mock
	mockUseCase := &MockWeatherUseCase{
		ExecuteFn: func(zipcode string) (map[string]float64, error) {
			return map[string]float64{
				"temp_C": 25.0,
				"temp_F": 77.0,
				"temp_K": 298.15,
			}, nil
		},
	}

	// Create handler with mock
	handler := &WeatherHandler{
		ServicoBUseCase: mockUseCase,
	}

	// Setup request
	req := httptest.NewRequest("GET", "/weather/12345678", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("zipcode", "12345678")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Setup response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ProcessServicoB(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Verify response body
	var response map[string]float64
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response body: %v", err)
	}

	expectedTemp := 25.0
	if response["temp_C"] != expectedTemp {
		t.Errorf("expected temp_C %f, got %f", expectedTemp, response["temp_C"])
	}

	// Verify mock was called with correct arguments
	if !mockUseCase.ExecuteCalled {
		t.Error("expected Execute method to be called")
	}
	if mockUseCase.ExecuteZipcode != "12345678" {
		t.Errorf("expected Execute to be called with zipcode %s, got %s", "12345678", mockUseCase.ExecuteZipcode)
	}
}

func TestWeatherHandler_GetWeather_Error(t *testing.T) {
	// Setup mock with error
	mockUseCase := &MockWeatherUseCase{
		ExecuteFn: func(zipcode string) (map[string]float64, error) {
			return nil, errors.New("invalid zipcode")
		},
	}

	// Create handler with mock
	handler := &WeatherHandler{
		ServicoBUseCase: mockUseCase,
	}

	// Setup request
	req := httptest.NewRequest("GET", "/weather/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("zipcode", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Setup response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ProcessServicoB(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify response body contains error message
	if w.Body.String() != "invalid zipcode\n" {
		t.Errorf("expected error message %q, got %q", "invalid zipcode\n", w.Body.String())
	}
}

func TestNewWeatherHandler(t *testing.T) {
	apiKey := "test-api-key"
	handler := NewWeatherHandler(apiKey)

	if handler == nil {
		t.Error("expected non-nil handler")
	}

	// Check that the handler has a use case
	if handler.ServicoBUseCase == nil {
		t.Error("expected non-nil ServicoBUseCase")
	}
}
