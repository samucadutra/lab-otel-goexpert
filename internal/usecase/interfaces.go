package usecase

// ServicoBWeatherUseCaseInterface defines the interface for getting weather by zipcode
type ServicoBWeatherUseCaseInterface interface {
	Execute(zipcode string) (map[string]float64, error)
}
