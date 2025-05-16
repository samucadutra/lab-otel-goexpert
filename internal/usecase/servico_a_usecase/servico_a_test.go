package servico_a_usecase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestNewServicoAUseCase(t *testing.T) {
	tests := []struct {
		name    string
		zipCode interface{}
		want    *ServicoAUseCase
	}{
		{
			name:    "should create a new service with string zipcode",
			zipCode: "12345678",
			want:    &ServicoAUseCase{ZipCode: "12345678"},
		},
		{
			name:    "should create a new service with integer zipcode",
			zipCode: 12345678,
			want:    &ServicoAUseCase{ZipCode: 12345678},
		},
		{
			name:    "should create a new service with nil zipcode",
			zipCode: nil,
			want:    &ServicoAUseCase{ZipCode: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServicoAUseCase(tt.zipCode)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServicoAUseCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidZipCode(t *testing.T) {
	tests := []struct {
		name    string
		zipCode string
		want    bool
	}{
		{
			name:    "should return true for valid 8-digit zipcode",
			zipCode: "12345678",
			want:    true,
		},
		{
			name:    "should return false for zipcode with less than 8 digits",
			zipCode: "1234567",
			want:    false,
		},
		{
			name:    "should return false for zipcode with more than 8 digits",
			zipCode: "123456789",
			want:    false,
		},
		{
			name:    "should return false for zipcode with non-digit characters",
			zipCode: "1234567A",
			want:    false,
		},
		{
			name:    "should return false for empty zipcode",
			zipCode: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidZipCode(tt.zipCode)
			if got != tt.want {
				t.Errorf("isValidZipCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServicoAUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/12345678":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"temp_c": 25.5, "temp_f": 77.9, "temp_k": 298.65}`))
		case "/99999999":
			w.WriteHeader(http.StatusNotFound)
		case "/88888888":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Set up viper config to use mock server
	viper.Set("EXTERNAL_CALL_URL", server.URL)

	tests := []struct {
		name        string
		zipCode     interface{}
		wantData    *WeatherData
		wantOk      bool
		wantErrMsg  string
		expectError bool
	}{
		{
			name:    "should succeed with valid zipcode",
			zipCode: "12345678",
			wantData: &WeatherData{
				TempC: 25.5,
				TempF: 77.9,
				TempK: 298.65,
			},
			wantOk:      true,
			expectError: false,
		},
		{
			name:        "should fail with invalid zipcode format",
			zipCode:     "1234567", // 7 digits
			wantData:    nil,
			wantOk:      false,
			wantErrMsg:  "invalid zipcode",
			expectError: true,
		},
		{
			name:        "should fail with non-string zipcode",
			zipCode:     12345678,
			wantData:    nil,
			wantOk:      false,
			wantErrMsg:  "invalid zipcode",
			expectError: true,
		},
		{
			name:        "should fail with nil zipcode",
			zipCode:     nil,
			wantData:    nil,
			wantOk:      false,
			wantErrMsg:  "invalid zipcode",
			expectError: true,
		},
		{
			name:        "should fail for not found zipcode",
			zipCode:     "99999999",
			wantData:    nil,
			wantOk:      true,
			wantErrMsg:  "can not find zipcode",
			expectError: true,
		},
		{
			name:        "should fail for server error",
			zipCode:     "88888888",
			wantData:    nil,
			wantOk:      true,
			wantErrMsg:  "failed to fetch weather data",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewServicoAUseCase(tt.zipCode)
			gotData, gotOk, err := uc.Execute(ctx)

			// Verify the ok result
			if gotOk != tt.wantOk {
				t.Errorf("Execute() gotOk = %v, want %v", gotOk, tt.wantOk)
			}

			// Verify error expectations
			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check error message when expected
			if err != nil && tt.wantErrMsg != "" {
				if err.Error() != tt.wantErrMsg && !contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Execute() error = %v, wantErr to contain %v", err, tt.wantErrMsg)
				}
				return
			}

			// Check result data
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("Execute() gotData = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || s[len(s)-len(substr):] == substr
}

func TestFetchCurrentWeather(t *testing.T) {
	ctx := context.Background()
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/12345678":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"temp_c": 25.5, "temp_f": 77.9, "temp_k": 298.65}`))
		case "/99999999":
			w.WriteHeader(http.StatusNotFound)
		case "/88888888":
			w.WriteHeader(http.StatusInternalServerError)
		case "/malformed":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{malformed json}`))
		}
	}))
	defer server.Close()

	// Set up viper config to use mock server
	viper.Set("EXTERNAL_CALL_URL", server.URL)

	tests := []struct {
		name        string
		zipCode     string
		wantData    *WeatherData
		expectError bool
		errorMsg    string
	}{
		{
			name:    "should return weather data for valid zipcode",
			zipCode: "12345678",
			wantData: &WeatherData{
				TempC: 25.5,
				TempF: 77.9,
				TempK: 298.65,
			},
			expectError: false,
		},
		{
			name:        "should return error for not found zipcode",
			zipCode:     "99999999",
			wantData:    nil,
			expectError: true,
			errorMsg:    "can not find zipcode",
		},
		{
			name:        "should return error for server error",
			zipCode:     "88888888",
			wantData:    nil,
			expectError: true,
			errorMsg:    "failed to fetch weather data",
		},
		{
			name:        "should return error for malformed JSON",
			zipCode:     "malformed",
			wantData:    nil,
			expectError: true,
			errorMsg:    "failed to decode weather data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := fetchCurrentWeather(ctx, tt.zipCode)

			if (err != nil) != tt.expectError {
				t.Errorf("fetchCurrentWeather() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if err != nil && tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
				t.Errorf("fetchCurrentWeather() error = %v, want to contain %v", err, tt.errorMsg)
				return
			}

			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("fetchCurrentWeather() = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}
