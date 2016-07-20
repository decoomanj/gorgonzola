package gorgonzola

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// holds the local address
var localAddress string

type MicroService struct {
	Admin   AdminServer
	Health  *Health
	Metrics *Metrics
	muxx    *mux.Router
}

type ContextHandler func(http.ResponseWriter, *http.Request, *Context)

// Set up the service
func init() {
	guessedAddress, err := guessLocalAddress()
	if err != nil {
		log.Fatal("No IP found! Exiting...")
	}
	log.Printf("Guess I'm running on %s\n", guessedAddress)
	localAddress = guessedAddress

}

// Instantiate a new microservice
func NewMicroService() *MicroService {
	return &MicroService{
		Admin:   NewAdminServer(),
		Health:  NewHealth(),
		Metrics: NewMetrics(),
		muxx:    mux.NewRouter(),
	}
}

// Wrap a Handler with AccessLogger and Principal
func (m *MicroService) Handle(method string, path string, handler ContextHandler) {
	log.Printf("Adding resource [%s] %s\n", method, path)
	m.muxx.Handle(path, Context{
		ms: m,
		next: HttpInstrument{
			AccessLogger{handler}.ServeHTTP,
		}.ServeHTTP,
	}).Methods(method)
}

// Wrap a Handler with AccessLogger and Principal
func (m *MicroService) Principal(method string, path string, handler ContextHandler) {
	fmt.Printf("Adding principal resource [%s] %s\n", method, path)
	m.muxx.Handle(path, Context{
		ms:   m,
		next: AccessLogger{Principal{handler}.ServeHTTP}.ServeHTTP,
	}).Methods(method)
}

// Handle: Not Allowed Requests
func (ms *MicroService) NotAllowed(method string, path string) {
	fmt.Printf("NotAllowed resource [%s] %s\n", method, path)
	MethodNotAllowed := func(w http.ResponseWriter, r *http.Request, c *Context) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	ms.muxx.Handle(path, Context{
		ms:   ms,
		next: AccessLogger{MethodNotAllowed}.ServeHTTP,
	}).Methods(method)
}

// Start a microservice with default health page on the given port
func (ms *MicroService) StartOnPort(port int) {

	// Start admin server
	go ms.Admin.Serve(ms)

	// start the web server
	log.Printf("MicroService is listening on %d....\n", port)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), ms.muxx); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// Start a microservice with default health page. It uses port 8080 by
// convention.
func (ms *MicroService) Start() {
	ms.StartOnPort(8080)
}
