# RedGo

### Настройка окружения

```bash
go mod tidy

```

### Запуск инфраструктуры

1. Запустите docker-compose:

```bash
docker-compose up -d

```

2. Выполните миграции (если таблицы не создались автоматически):

```bash
docker run --rm -v $(pwd)/migrate:/migrations --network host migrate/migrate \
    -path=/migrations/ -database "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable" up

```

**Заглушка (Stub-приемник)**
Запускает локальный сервер на порту `:9090` и встроенный ngrok-туннель:

```bash
go run cmd/stub/main.go

```

Если надо зачем то отдельно ngrok запустить то:

```bash
ngrok http 9090 --url=https://consuelo-extralegal-ray.ngrok-free.dev --authtoken=38FPlzBKW8qIYa07PAugdRjnfKR_5D6okQacPLbBydFwr8mSt

```

---

### Примеры запросов (API)

**Коллекция postman с запросами в [JSON](postman.json)**

**Коллекция с автоматически рандомными созданием инцидента и проверка локации: [JSON](nagruzka.json)**
<details>
<summary>Создание инцидента (POST)</summary>

![createIncident.png](photo/createIncident.png)
</details>

<details>
<summary>Получение статистики (GET)</summary>

![getStats.png](photo/getStats.png)
</details>

<details>
<summary>Получение инцидента (GET)</summary>

![getIncidents.png](photo/getIncidents.png)
</details>

<details>
<summary>Здоровье (GET)</summary>

![healthCheck.png](photo/healthCheck.png)
</details>

<details>
<summary>Проверка локации (POST)</summary>

![checkLocation.png](photo/checkLocation.png)
</details>

---

#### Логи приложения и воркеров
```log
2026/01/17 18:33:11 Successfully sent webhook for incident 72 (User 364)

2026/01/17 18:33:14 Scheduler: moved 3 tasks to main queue

2026/01/17 18:33:14 Successfully sent webhook for incident 116 (User 10)

2026/01/17 18:33:14 Task 57 scheduled for retry #4 in 16s

2026/01/17 18:33:15 Scheduler: moved 2 tasks to main queue

2026/01/17 18:33:15 Successfully sent webhook for incident 45 (User 337)

2026/01/17 18:33:15 Successfully sent webhook for incident 57 (User 356)

2026/01/17 18:33:16 Scheduler: moved 1 tasks to main queue

2026/01/17 18:33:16 Successfully sent webhook for incident 129 (User 338)

2026/01/17 18:33:19 Task 163 scheduled for retry #4 in 16s

2026/01/17 18:33:30 Scheduler: moved 1 tasks to main queue

2026/01/17 18:33:30 Successfully sent webhook for incident 57 (User 185)

2026/01/17 18:33:35 Scheduler: moved 1 tasks to main queue

2026/01/17 18:33:35 Successfully sent webhook for incident 163 (User 123)

[GIN-debug] redirecting request 301: /api/v1/incidents/ --> /api/v1/incidents/

[GIN-debug] redirecting request 307: /api/v1/incidents/ --> /api/v1/incidents/

```

#### Работа Ngrok и Заглушки
```log
Подключаемся к ngrok
localhost:9090
URL: https://consuelo-extralegal-ray.ngrok-free.dev
Имитация зависания
Ответ: 200
Ответ: 200
Ответ: 200
Ответ: 200
Ответ: 500
Имитация зависания
Имитация зависания
Ответ: 200
```

#### Логи выхода приложения когда нагружены воркеры
```log
2026/01/17 20:27:14 Successfully sent webhook for incident 145 (User 127)

2026/01/17 20:27:14 Successfully sent webhook for incident 145 (User 3)

2026/01/17 20:27:14 Task 75 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Stopping scheduler...

2026/01/17 20:27:14 Shutting down gracefully...

2026/01/17 20:27:14 Task 145 scheduled for retry #2 in 4s

2026/01/17 20:27:14 Worker #9 stopped

2026/01/17 20:27:14 Task 138 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Worker #7 stopped

2026/01/17 20:27:14 Task 110 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Worker #6 stopped

2026/01/17 20:27:14 Task 54 scheduled for retry #1 in 2s

2026/01/17 20:27:14 Worker #2 stopped

2026/01/17 20:27:14 Task 83 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Worker #10 stopped

2026/01/17 20:27:14 Task 145 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Worker #8 stopped

2026/01/17 20:27:14 Waiting for workers to finish current tasks...

2026/01/17 20:27:14 Task 182 scheduled for retry #1 in 2s

2026/01/17 20:27:14 Worker #4 stopped

2026/01/17 20:27:14 Task 59 scheduled for retry #1 in 2s

2026/01/17 20:27:14 Worker #1 stopped

2026/01/17 20:27:14 Task 131 scheduled for retry #3 in 8s

2026/01/17 20:27:14 Worker #5 stopped

2026/01/17 20:27:14 Task 94 scheduled for retry #1 in 2s

2026/01/17 20:27:14 Worker #3 stopped

2026/01/17 20:27:14 All workers exited cleanly

Server exited properly
```

---