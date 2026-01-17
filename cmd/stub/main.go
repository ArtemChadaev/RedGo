package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.ngrok.com/ngrok/v2"
)

func main() {
	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден")
	}

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	token := os.Getenv("NGROK_AUTHTOKEN")
	domain := os.Getenv("NGROK_DOMAIN")

	if token == "" || domain == "" {
		return fmt.Errorf("критическая ошибка: NGROK_AUTHTOKEN или NGROK_DOMAIN не заданы в .env")
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		chance := rand.Intn(100)
		switch {
		case chance < 10:
			fmt.Println("Имитация зависания")
			<-make(chan struct{})

		case chance < 20:
			fmt.Println("Ответ: 500")
			w.WriteHeader(http.StatusInternalServerError)

		default:
			fmt.Println("Ответ: 200")
			w.WriteHeader(http.StatusOK)
		}
	})

	// Фоновое прослушивание 9090
	go func() {
		localAddr := ":9090"
		fmt.Println("localhost" + localAddr)
		if err := http.ListenAndServe(localAddr, handler); err != nil {
			log.Printf("Ошибка локального сервера: %v", err)
		}
	}()

	fmt.Println("Подключаемся к ngrok")
	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(token))
	if err != nil {
		return fmt.Errorf("Ошибка создания агента: %w", err)
	}

	ln, err := agent.Listen(ctx, ngrok.WithURL(domain))
	if err != nil {
		return fmt.Errorf("Ошибка Listen: %w", err)
	}

	fmt.Println("URL:", ln.URL())

	return http.Serve(ln, handler)
}
