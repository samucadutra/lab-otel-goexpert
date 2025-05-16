package servico_b_usecase

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewServicoBUseCase(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   *ServicoBUseCase
	}{
		{
			name:   "should create a new service with API key",
			apiKey: "test-api-key",
			want:   &ServicoBUseCase{WeatherApiKey: "test-api-key"},
		},
		{
			name:   "should create a new service with empty API key",
			apiKey: "",
			want:   &ServicoBUseCase{WeatherApiKey: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServicoBUseCase(tt.apiKey)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServicoBUseCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidZipCode(t *testing.T) {
	tests := []struct {
		name    string
		zipcode string
		want    bool
	}{
		{
			name:    "should return true for valid 8-digit zipcode",
			zipcode: "12345678",
			want:    true,
		},
		{
			name:    "should return false for zipcode with less than 8 digits",
			zipcode: "1234567",
			want:    false,
		},
		{
			name:    "should return false for zipcode with more than 8 digits",
			zipcode: "123456789",
			want:    false,
		},
		{
			name:    "should return false for zipcode with non-digit characters",
			zipcode: "1234567A",
			want:    false,
		},
		{
			name:    "should return false for empty zipcode",
			zipcode: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidZipCodeFn(tt.zipcode)
			if got != tt.want {
				t.Errorf("isValidZipCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchLocation(t *testing.T) {
	// Create a test server for viacep API
	locationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zipcode := r.URL.Path[len("/ws/") : len(r.URL.Path)-len("/json/")]

		switch zipcode {
		case "12345678":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"localidade": "São Paulo"}`))
		case "99999999":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"localidade": ""}`))
		case "88888888":
			w.WriteHeader(http.StatusNotFound)
		case "77777777":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer locationServer.Close()

	// Override the viacep URL for testing
	originalGet := httpGet
	httpGet = func(url string) (*http.Response, error) {
		newURL := locationServer.URL + url[len("https://viacep.com.br"):]
		return http.DefaultClient.Get(newURL)
	}
	defer func() { httpGet = originalGet }()

	tests := []struct {
		name        string
		zipcode     string
		want        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "should return location for valid zipcode",
			zipcode:     "12345678",
			want:        "São Paulo",
			expectError: false,
		},
		{
			name:        "should return error for empty location",
			zipcode:     "99999999",
			want:        "",
			expectError: true,
			errorMsg:    "can not find zipcode",
		},
		{
			name:        "should return error for not found zipcode",
			zipcode:     "88888888",
			want:        "",
			expectError: true,
			errorMsg:    "failed to fetch location",
		},
		{
			name:        "should return error for invalid JSON",
			zipcode:     "77777777",
			want:        "",
			expectError: true,
			errorMsg:    "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchLocationFn(tt.zipcode)

			if (err != nil) != tt.expectError {
				t.Errorf("fetchLocation() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("fetchLocation() error = %v, want error containing %v", err, tt.errorMsg)
				}
				return
			}

			if got != tt.want {
				t.Errorf("fetchLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchWeather(t *testing.T) {
	// Create a test server for weather API
	weatherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		location := r.URL.Query().Get("q")
		apiKey := r.URL.Query().Get("key")

		if apiKey != "valid-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch location {
		case "São Paulo":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"current": {"temp_c": 25.5, "temp_f": 77.9}}`))
		case "Rio de Janeiro":
			w.WriteHeader(http.StatusInternalServerError)
		case "Curitiba":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer weatherServer.Close()

	// Save the original NewRequest and DefaultClient.Do functions
	originalNewRequest := httpNewRequest
	originalDo := httpClientDo

	// Mock the functions for testing
	httpNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		// Use the test server URL instead of the original URL
		req, err := http.NewRequest(method, weatherServer.URL, body)
		if err != nil {
			return nil, err
		}
		return req, nil
	}

	httpClientDo = func(req *http.Request) (*http.Response, error) {
		// We don't need to modify the URL here, just use the standard client
		return http.DefaultClient.Do(req)
	}

	// Restore the original functions after the test
	defer func() {
		httpNewRequest = originalNewRequest
		httpClientDo = originalDo
	}()

	tests := []struct {
		name        string
		location    string
		apiKey      string
		want        *WeatherData
		expectError bool
		errorMsg    string
	}{
		{
			name:     "should return weather data for valid location and API key",
			location: "São Paulo",
			apiKey:   "valid-key",
			want: &WeatherData{
				TempC: 25.5,
				TempF: 77.9,
			},
			expectError: false,
		},
		{
			name:        "should return error for invalid API key",
			location:    "São Paulo",
			apiKey:      "invalid-key",
			want:        nil,
			expectError: true,
			errorMsg:    "failed to fetch weather data",
		},
		{
			name:        "should return error for server error",
			location:    "Rio de Janeiro",
			apiKey:      "valid-key",
			want:        nil,
			expectError: true,
			errorMsg:    "failed to fetch weather data",
		},
		{
			name:        "should return error for invalid JSON",
			location:    "Curitiba",
			apiKey:      "valid-key",
			want:        nil,
			expectError: true,
			errorMsg:    "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchWeatherFn(tt.location, tt.apiKey)

			if (err != nil) != tt.expectError {
				t.Errorf("fetchWeather() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("fetchWeather() error = %v, want error containing %v", err, tt.errorMsg)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchWeather() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServicoBUseCase_Execute(t *testing.T) {
	// Save original functions to restore later
	originalIsValidZipCode := isValidZipCodeFn
	originalFetchLocation := fetchLocationFn
	originalFetchWeather := fetchWeatherFn

	// Restore original functions after tests
	defer func() {
		isValidZipCodeFn = originalIsValidZipCode
		fetchLocationFn = originalFetchLocation
		fetchWeatherFn = originalFetchWeather
	}()

	tests := []struct {
		name             string
		zipcode          string
		apiKey           string
		mockValidZipCode bool
		mockLocation     string
		mockLocErr       error
		mockWeather      *WeatherData
		mockWeatherErr   error
		want             map[string]float64
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "should return weather data for valid inputs",
			zipcode:          "12345678",
			apiKey:           "valid-key",
			mockValidZipCode: true,
			mockLocation:     "São Paulo",
			mockLocErr:       nil,
			mockWeather:      &WeatherData{TempC: 25.5, TempF: 77.9},
			mockWeatherErr:   nil,
			want: map[string]float64{
				"temp_C": 25.5,
				"temp_F": 77.9,
				"temp_K": 298.65,
			},
			expectError: false,
		},
		{
			name:             "should return error for invalid zipcode",
			zipcode:          "invalid",
			apiKey:           "valid-key",
			mockValidZipCode: false,
			want:             nil,
			expectError:      true,
			errorMsg:         "invalid zipcode",
		},
		{
			name:             "should return error from location fetch",
			zipcode:          "12345678",
			apiKey:           "valid-key",
			mockValidZipCode: true,
			mockLocation:     "",
			mockLocErr:       fmt.Errorf("can not find zipcode"),
			want:             nil,
			expectError:      true,
			errorMsg:         "can not find zipcode",
		},
		{
			name:             "should return error from weather fetch",
			zipcode:          "12345678",
			apiKey:           "valid-key",
			mockValidZipCode: true,
			mockLocation:     "São Paulo",
			mockLocErr:       nil,
			mockWeather:      nil,
			mockWeatherErr:   fmt.Errorf("failed to fetch weather data"),
			want:             nil,
			expectError:      true,
			errorMsg:         "failed to fetch weather data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock dependencies
			isValidZipCodeFn = func(zipcode string) bool {
				return tt.mockValidZipCode
			}

			fetchLocationFn = func(zipcode string) (string, error) {
				return tt.mockLocation, tt.mockLocErr
			}

			fetchWeatherFn = func(location, apiKey string) (*WeatherData, error) {
				return tt.mockWeather, tt.mockWeatherErr
			}

			uc := NewServicoBUseCase(tt.apiKey)
			got, err := uc.Execute(tt.zipcode)

			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errorMsg)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains another string
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || (len(s) > len(substr) && s[1:len(s)-1] == substr))
}
