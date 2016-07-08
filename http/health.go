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
type HealthCheck func() error

type HealthStatus struct {
	sync.RWMutex
	items map[string]string
}

type HealthChecks struct {
	sync.RWMutex
	items map[string]HealthCheck
}

type Health struct {
	checks HealthChecks
	status HealthStatus
}

type Status struct {
	Status string            `json:"status"`
	Errors map[string]string `json:"errors,omitempty"`
}

func NewHealth() Health {
	return Health{
		checks: HealthChecks{items: make(map[string]HealthCheck)},
		status: HealthStatus{items: make(map[string]string)},
	}
}

// Register a healthcheck. The health-check should not block and may not take
// longer than 1s to finish.
func (hc *Health) Register(name string, healthCheck HealthCheck) {
	log.Printf("Registering Health Check: %s\n", name)
	hc.checks.items[name] = healthCheck

}

// aggregateState collects the state of the application
func (health *Health) aggregateState() {
	for {
		var wg sync.WaitGroup
		health.checks.RLock()
		defer health.checks.RUnlock()
		for name, x := range health.checks.items {
			wg.Add(1)
			go func(name string, healthCheck HealthCheck) {
				defer wg.Done()
				if err := timeout(healthCheck); err == nil {
					health.status.remove(name)
				} else {
					health.status.add(name, err.Error())
				}
			}(name, x)
		}
		wg.Wait()
		time.Sleep(time.Second)
	}
}

// remove a check from the status list
func (hs *HealthStatus) remove(name string) {
	hs.Lock()
	defer hs.Unlock()
	delete(hs.items, name)
}

// add a check to the status list
func (hs *HealthStatus) add(name string, message string) {
	hs.Lock()
	defer hs.Unlock()
	hs.items[name] = message
}

// timeout wait for the healthcheck function to return. After 1s the timeout
// is thrown.
func timeout(healthCheck HealthCheck) error {
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
	h.status.RLock()
	defer h.status.RUnlock()

	if len(h.status.items) == 0 {
		writeStatus(w, Status{"up", nil}, http.StatusOK)
	} else {
		writeStatus(w, Status{"down", h.status.items}, http.StatusServiceUnavailable)
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
