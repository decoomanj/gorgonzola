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
func Do(w http.ResponseWriter, r *http.Request, c *gorgonzola.Context) {

	fmt.Println("PROCESSING")

	sha_256 := sha256.New()
	io.Copy(sha_256, r.Body)
	//sha_256.Write(r.Body.Read())
	fmt.Printf("sha256:\t%x\n", sha_256.Sum(nil))

	// set status
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "some content")

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

	ms := gorgonzola.NewMicroService()
	ms.Health.Register(a)
	ms.Health.Register(b)

	ms.Metrics.Register(&gorgonzola.MMetric{Name: "memory.alloc"}, time.Millisecond*500)

	ms.Handle("POST", "/jan", Do)

	ms.Start()

	fmt.Println("Stopping")
}
