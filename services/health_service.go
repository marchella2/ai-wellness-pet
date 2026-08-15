package services

// HealthService holds the health check business logic.
type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

// HealthStatus is the health check response payload.
type HealthStatus struct {
	Status string `json:"status"`
}

// Check returns the current server status.
func (s *HealthService) Check() HealthStatus {
	return HealthStatus{Status: "ok"}
}
