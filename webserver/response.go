package webserver

type SettingsResponse struct {
	APIEndpoint string `json:"api_endpoint"`
}

//HealthResponse is returned when healtcheck requested
type HealthResponse struct {
	Status bool `json:"status"`
}
