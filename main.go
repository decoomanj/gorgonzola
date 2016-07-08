package main

import (
	//"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"./http"
)

// Delete a file from the storage
func Do(w http.ResponseWriter, r *http.Request, c *gorgonzola.Context) {

	// set status
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "some content")

}

func myFunc() error {
	//fmt.Println("HIER")
	time.Sleep(time.Second * 5)
	return nil
}

func myFunc2() error {
	//fmt.Println("DAAR")
	//return errors.New("voilá")
	return nil
}

func main() {

	log.Println("MicroService Showcase")

	ms := gorgonzola.NewMicroService()
	ms.Admin.Health.Register("test", myFunc)
	ms.Admin.Health.Register("test2", myFunc2)

	ms.Handle("GET", "/jan", Do)

	ms.Start()

	fmt.Println("Stopping")
}
