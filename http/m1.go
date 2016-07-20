package gorgonzola

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
)

type HttpInstrument struct {
	next ContextHandler
}

// Wrap the request with access logging
func (p HttpInstrument) ServeHTTP(w http.ResponseWriter, r *http.Request, c *Context) {

	x := c.ms.Metrics.GetOrCreate("http.out.edge", func(name string) XMetric { return NewValveVec(name) }).(*ValveVec)
	remoteAddr, _ := stripAddress(r.RemoteAddr)

	r.Body = RequestMetrics{r.Body, func(v float64) { x.Add(remoteAddr, v) }}
	respDel := &responseWriterDelegator{ResponseWriter: w}
	p.next(respDel, r, c)
	log.Println(respDel.status, respDel.written)

}

type RequestMetrics struct {
	io.ReadCloser
	t func(v float64)
}

// Observe how many bytes are sent in the request
func (l RequestMetrics) Read(p []byte) (int, error) {
	inBytes, err := l.ReadCloser.Read(p)
	l.t(float64(inBytes))
	return inBytes, err
}

// We need this to remember the status code
type responseWriterDelegator struct {
	http.ResponseWriter
	written     int64
	status      int
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

// Register.Gauge("http") -> Label(ip, "in") xx

//    ms.Metrics.Register( Valve("ip", "in") )
//    ms.Metrics.Register( Counter("ip", "in") )
//    ms.Metrics.Register("ip", "out").Add()
//	  ms.Metrics.Register("proc", "cpu", func())
//    ms.Metrics.All.Reap() -> timestamp
//    ms.Metrics.Get("ip", "in").(*Valve).Add()
//    ms.Metrics.Get("ip", "in").(*Counter).Inc()

// Guess the local IP
func guessLocalAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("No IP found!")
}

// Map the localhost (IPv4, IPv6) address to the exposed IP
func remap(address string) string {
	if address == "127.0.0.1" || address == "::1" {
		return localAddress
	}
	return address
}

// Clear the port information from ip:port
func stripAddress(remoteAddress string) (string, error) {
	ip, _, err := net.SplitHostPort(remoteAddress)
	return remap(ip), err
}
