package api

import "net/url"

//HealthResponse is returned when healtcheck requested
type HealthResponse struct {
	Status bool `json:"status"`
}

//HttpErrorResponse is returned on error
type HttpErrorResponse struct {
	Error      string     `json:"error"`
	ErrorQuery url.Values `json:"errorQuery"`
}
