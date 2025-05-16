package servico_b_usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

var (
	httpGet        = http.Get
	httpNewRequest = http.NewRequest
	httpClientDo   = http.DefaultClient.Do

	isValidZipCodeFn = isValidZipCodeImpl
	fetchLocationFn  = fetchLocationImpl
	fetchWeatherFn   = fetchWeatherImpl
)

type WeatherData struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
}

type ServicoBUseCase struct {
	WeatherApiKey string
}

func NewServicoBUseCase(apiKey string) *ServicoBUseCase {
	return &ServicoBUseCase{WeatherApiKey: apiKey}
}

func (uc *ServicoBUseCase) Execute(zipcode string) (map[string]float64, error) {
	if !isValidZipCodeFn(zipcode) {
		return nil, fmt.Errorf("invalid zipcode")
	}

	location, err := fetchLocationFn(zipcode)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}

	weather, err := fetchWeatherFn(location, uc.WeatherApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}

	tempK := weather.TempC + 273.15

	return map[string]float64{
		"temp_C": weather.TempC,
		"temp_F": weather.TempF,
		"temp_K": tempK,
	}, nil
}

func isValidZipCodeImpl(zipcode string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(zipcode)
}

func fetchLocationImpl(zipcode string) (string, error) {
	resp, err := httpGet(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", zipcode))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch location")
	}

	var data struct {
		Localidade string `json:"localidade"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.Localidade == "" {
		return "", fmt.Errorf("can not find zipcode")
	}

	return data.Localidade, nil
}

func fetchWeatherImpl(location, apiKey string) (*WeatherData, error) {
	baseURL := "http://api.weatherapi.com/v1/current.json"
	req, err := httpNewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", apiKey)
	q.Add("q", location)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClientDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch weather data")
	}

	var data struct {
		Current WeatherData `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data.Current, nil
}
