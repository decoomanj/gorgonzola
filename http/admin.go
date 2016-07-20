package gorgonzola

import (
	"log"
	"net/http"
	"strconv"

	//"github.com/prometheus/client_golang/prometheus"
)

// AdminMux is the server to handle administrative requests
type AdminServer struct {
	muxx *http.ServeMux
	ms   *MicroService
}

func NewAdminServer() AdminServer {
	return AdminServer{}
}

func (admin *AdminServer) Serve(ms *MicroService) {

	log.Println("Starting administration server")
	muxx := http.NewServeMux()

	// Start reaping health checks
	//go admin.Metrics.aggregate()

	// add health
	muxx.HandleFunc("/health", ms.Health.page)

	// add metrics
	muxx.HandleFunc("/metrics", ms.Metrics.page)
	//admin.muxx.Handle("/metrics", prometheus.Handler())

	// default port
	port := 9090

	// start the web server
	log.Printf("Administrator is listening on %d....\n", port)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), muxx); err != nil {
		log.Fatal("Administrator ListenAndServe:", err)
	}
}
