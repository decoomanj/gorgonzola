package gorgonzola

import (
	"io"
	"log"
	"net/http"
)

type HttpInstrument struct {
	next ContextHandler
}

// idee: create "key" = hash(name+ label-KV)
// update value there: no matrix needed
// id := sprintf("http_edge_in,from=%s,ip=%s,port=%s"...)
func CreateHttpValve(name string) Metric { return NewValveVec(name) }

// Wrap the request with access logging
func (p HttpInstrument) ServeHTTP(w http.ResponseWriter, r *http.Request, c *Context) {

	x := c.ms.Metrics.GetOrCreate("http_edge_in", CreateHttpValve).(*ValveVec)
	y := c.ms.Metrics.GetOrCreate("http_edge_out", CreateHttpValve).(*ValveVec)
	remoteAddr, _ := stripAddress(r.RemoteAddr)

	r.Body = requestDecorator{r.Body, func(v float64) { x.Add(remoteAddr, v) }}
	respDel := &responseDelegator{ResponseWriter: w, metric: func(v float64) { y.Add(remoteAddr, v) }}
	p.next(respDel, r, c)

	log.Println(respDel.status)
}

type requestDecorator struct {
	io.ReadCloser
	metric func(v float64)
}

// Observe how many bytes are sent in the request
func (l requestDecorator) Read(p []byte) (int, error) {
	inBytes, err := l.ReadCloser.Read(p)
	l.metric(float64(inBytes))
	return inBytes, err
}

// We need this to remember the status code
type responseDelegator struct {
	http.ResponseWriter
	metric      func(v float64)
	status      int
	wroteHeader bool
}

func (r *responseDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.metric(float64(n))
	return n, err
}
