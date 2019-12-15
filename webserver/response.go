package webserver

//HealthResponse is returned when healtcheck requested
type HealthResponse struct {
	Status bool `json:"status"`
}

//HttpErrorResponse is returned on error
type HttpErrorResponse struct {
	Error string `json:"error"`
}
