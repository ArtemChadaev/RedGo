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
	// Загружаем переменные из .env (NGROK_AUTHTOKEN и NGROK_DOMAIN)
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

	// Проверяем, что переменные вообще дошли до программы
	fmt.Printf("DEBUG: Token length: %d\n", len(token))
	fmt.Printf("DEBUG: Domain: %s\n", domain)

	if token == "" || domain == "" {
		return fmt.Errorf("критическая ошибка: NGROK_AUTHTOKEN или NGROK_DOMAIN не заданы в .env")
	}

	fmt.Println("⏳ Подключаемся к ngrok... (это может занять до 10 секунд)")

	agent, err := ngrok.NewAgent(
		ngrok.WithAuthtoken(token),
	)
	if err != nil {
		return fmt.Errorf("ошибка создания агента: %w", err)
	}

	// Здесь программа может висеть, если домен занят или сеть тупит
	ln, err := agent.Listen(ctx,
		ngrok.WithURL(domain),
	)
	if err != nil {
		return fmt.Errorf("ошибка Listen: %w", err)
	}

	fmt.Println("🚀 Малыш-приемник запущен на постоянном адресе:", ln.URL())

	// 3. Наш обработчик с твоей логикой
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		chance := rand.Intn(100)
		switch {
		case chance < 10:
			fmt.Println("⚠️ Имитация зависания (Hanging...)")
			<-make(chan struct{})

		case chance < 20:
			fmt.Println("❌ Ответ: 500 Internal Server Error")
			w.WriteHeader(http.StatusInternalServerError)

		default:
			fmt.Println("✅ Ответ: 200 OK")
			w.WriteHeader(http.StatusOK)
		}
	})

	// Запускаем сервер прямо на туннеле ngrok
	return http.Serve(ln, handler)
}
