package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	port := ":9090"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		chance := rand.Intn(100)

		switch {
		case chance < 20: // 20% шанс на долгий ответ
			fmt.Println("Timeout")
			time.Sleep(15 * time.Second)
			w.WriteHeader(http.StatusOK)

		case chance < 40: // 20% шанс на внутреннюю ошибку сервера
			fmt.Println("500 Internal Server Error")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "portal_crashed"}`))

		default:
			fmt.Printf("✅ Успешно получено: %s\n", string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
