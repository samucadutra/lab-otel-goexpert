package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/samucadutra/lab-otel-goexpert/internal/usecase"
	"github.com/samucadutra/lab-otel-goexpert/internal/usecase/servico_a_usecase"
	"github.com/samucadutra/lab-otel-goexpert/internal/usecase/servico_b_usecase"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type WeatherHandler struct {
	ServicoBUseCase usecase.ServicoBWeatherUseCaseInterface
	tracer          trace.Tracer
}

type WeatherRequest struct {
	ZipCode interface{} `json:"cep"`
}

func NewWeatherHandler(tracer trace.Tracer) *WeatherHandler {
	apiKey := viper.GetString("WEATHER_API_KEY")
	return &WeatherHandler{
		ServicoBUseCase: servico_b_usecase.NewServicoBUseCase(apiKey),
		tracer:          tracer,
	}
}

func (h *WeatherHandler) ProcessServicoA(w http.ResponseWriter, r *http.Request) {
	var request WeatherRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	servicoAUC := servico_a_usecase.NewServicoAUseCase(request.ZipCode)

	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := h.tracer.Start(ctx, viper.GetString("REQUEST_NAME_OTEL"))
	defer span.End()

	response, isValid, err := servicoAUC.Execute(ctx)
	if err != nil && !isValid {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		if err.Error() == "can not find zipcode" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *WeatherHandler) ProcessServicoB(w http.ResponseWriter, r *http.Request) {
	zipcode := chi.URLParam(r, "zipcode")

	response, err := h.ServicoBUseCase.Execute(zipcode)
	if err != nil {
		if err.Error() == "can not find zipcode" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
