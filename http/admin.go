package gorgonzola

import (
	"log"
	"net/http"
	"strconv"
)

// AdminMux is the server to handle administrative requests
type AdminServer struct {
	muxx   *http.ServeMux
	Health Health
}

func NewAdminServer() AdminServer {
	return AdminServer{
		muxx:   http.NewServeMux(),
		Health: NewHealth(),
	}
}

func (admin *AdminServer) Serve() {
	log.Println("Starting administration server")

	// Start reaping health checks
	go admin.Health.aggregateState()

	// add health
	admin.muxx.HandleFunc("/health", admin.Health.page)

	// add metrics

	// default port
	port := 9090

	// start the web server
	log.Printf("Administrator is listening on %d....\n", port)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), admin.muxx); err != nil {
		log.Fatal("Administrator ListenAndServe:", err)
	}
}
