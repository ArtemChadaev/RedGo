package main

import (
	"log"
	"math/rand"
	"net/http"
)

func main() {
	port := ":9090"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		chance := rand.Intn(100)

		switch {
		case chance < 10:
			<-make(chan struct{})

		case chance < 20:
			w.WriteHeader(http.StatusInternalServerError)

		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
