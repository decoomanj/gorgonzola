package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"crypto/sha256"

	gorgonzola "./http"
)

// Delete a file from the storage
func Post(w http.ResponseWriter, r *http.Request, c *gorgonzola.Context) {

	fmt.Println("PROCESSING")

	sha_256 := sha256.New()
	io.Copy(sha_256, r.Body)
	//sha_256.Write(r.Body.Read())
	fmt.Printf("sha256:\t%x\n", sha_256.Sum(nil))

	// set status
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "some content")

}

// Delete a file from the storage
func Get(w http.ResponseWriter, r *http.Request, c *gorgonzola.Context) {

	// set status
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("test"))
}

func myFunc() error {
	//fmt.Println("HIER")
	//time.Sleep(time.Second * 6)
	return nil
}

func myFunc2() error {
	//fmt.Println("DAAR")
	//return errors.New("voilá")
	return nil
}

func myMetric() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(600) + 1
}

func main() {

	log.Println("MicroService Showcase")

	a := &gorgonzola.HealthCheck{Name: "test", Handler: myFunc, Interval: time.Millisecond * 100}
	b := &gorgonzola.HealthCheck{Name: "test2", Handler: myFunc2, Interval: time.Second * 2}

	ms := gorgonzola.NewMicroService("test-service")
	ms.Health.Register(a)
	ms.Health.Register(b)

	ms.Metrics.Register(&gorgonzola.GoMetrics{}, time.Millisecond*500)

	ms.Handle("POST", "/jan", Post)
	ms.Handle("GET", "/jan", Get)

	ms.Start()

	fmt.Println("Stopping")
}
