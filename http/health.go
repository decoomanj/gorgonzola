package gorgonzola

import (
	"encoding/json"
	"errors"

	"log"
	"net/http"
	"sync"
	"time"
)

// Health
type HealthFunc func() error

type HealthCheck struct {
	Name     string
	Handler  HealthFunc
	Interval time.Duration
	healthy  bool
	msg      string
}

type HealthChecks struct {
	sync.RWMutex
	items map[string]*HealthCheck
}

type Health struct {
	checks HealthChecks
}

type Status struct {
	Status string            `json:"status"`
	Errors map[string]string `json:"errors,omitempty"`
}

// NewHealth creates a new Health struct
func NewHealth() *Health {
	return &Health{
		checks: NewHealthChecks(),
	}
}

// NewHealthChekcs creates a new HealthChecks struct
func NewHealthChecks() HealthChecks {
	return HealthChecks{items: make(map[string]*HealthCheck)}
}

// Register a healthcheck. The health-check should not block and may not take
// longer than 1s to finish.
func (hc *Health) Register(healthCheck *HealthCheck) {
	log.Printf("Registering Health Check: %s\n", healthCheck.Name)
	hc.checks.items[healthCheck.Name] = healthCheck
	go healthCheck.start()
}

// start checks the health function
func (hc *HealthCheck) start() {
	for {
		if err := timeout(hc.Handler); err == nil {
			hc.healthy = true
			hc.msg = ""
		} else {
			hc.healthy = false
			hc.msg = err.Error()
		}
		time.Sleep(hc.Interval)
	}
}

// timeout wait for the healthcheck function to return. After 1s the timeout
// is thrown.
func timeout(healthCheck HealthFunc) error {
	healthWait := make(chan error, 1)
	go func() {
		healthWait <- healthCheck()
	}()

	select {
	case res := <-healthWait:
		return res
	case <-time.After(time.Second):
		return errors.New("timeout")
	}
}

// page renders the health status page
func (h *Health) page(w http.ResponseWriter, r *http.Request) {
	h.checks.RLock()
	defer h.checks.RUnlock()

	errors := make(map[string]string)
	for name, hc := range h.checks.items {
		if !hc.healthy {
			errors[name] = hc.msg
		}
	}

	if len(errors) == 0 {
		writeStatus(w, Status{"up", nil}, http.StatusOK)
	} else {
		writeStatus(w, Status{"down", errors}, http.StatusServiceUnavailable)
	}

}

func writeStatus(w http.ResponseWriter, status Status, code int) {
	js, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(js)
}
