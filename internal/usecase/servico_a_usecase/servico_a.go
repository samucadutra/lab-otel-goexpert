package servico_a_usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net/http"
	"regexp"
)

type ServicoAUseCase struct {
	ZipCode interface{}
}

type WeatherData struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

func NewServicoAUseCase(zipcode interface{}) *ServicoAUseCase {
	return &ServicoAUseCase{
		ZipCode: zipcode,
	}
}

func (uc *ServicoAUseCase) Execute(ctx context.Context) (*WeatherData, bool, error) {
	zipCodeStr, ok := uc.ZipCode.(string)
	if !ok {
		log.Println("ZipCode is not a string")
		return nil, false, fmt.Errorf("invalid zipcode")
	}

	if !isValidZipCode(zipCodeStr) {
		log.Println("Invalid ZipCode format")
		return nil, false, fmt.Errorf("invalid zipcode")
	}

	weatherData, err := fetchCurrentWeather(ctx, zipCodeStr)
	if err != nil {
		return nil, true, fmt.Errorf("%s", err.Error())
	}

	return weatherData, true, nil
}

func fetchCurrentWeather(ctx context.Context, zipcode string) (*WeatherData, error) {
	externalCallURL := viper.GetString("EXTERNAL_CALL_URL")
	var req *http.Request
	var err error
	req, err = http.NewRequestWithContext(ctx, "GET", externalCallURL+"/"+zipcode, nil)
	//resp, err := http.Get(externalCallURL + "/" + zipcode)
	if err != nil {
		return nil, fmt.Errorf(err.Error(), err)
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("can not find zipcode")
		}
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	var weatherData WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return &weatherData, nil
}
func isValidZipCode(zipcode string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(zipcode)
}
