package contract

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environement"`
	Version     string `json:"version"`
}
