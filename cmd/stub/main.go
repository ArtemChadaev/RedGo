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
		case chance < 10: // 10% шанс "зависнуть" без ответа
			<-make(chan struct{})

		case chance < 20: // 10% шанс на внутреннюю ошибку сервера
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "portal_crashed"}`))

		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
