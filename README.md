# RedGo

В docker-compose должно импортироваться автоматически sql, но если что можно и через миграцию.
Для запуска миграции использовал:
```cmd
docker run --rm -v $(pwd)/migrate:/migrations --network host migrate/migrate \
    -path=/migrations/ -database "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable" up
```