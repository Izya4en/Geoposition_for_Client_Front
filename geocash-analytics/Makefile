# Переменные
APP_NAME=atm-service
# Строка подключения для локальных миграций (если запускаете migrate локально)
DB_URL=postgres://postgres:secret@localhost:5432/atm_db?sslmode=disable

.PHONY: help run build up down logs db-status clean

# --- Основные команды ---

# Запуск всего окружения в Docker (БД + Приложение)
up:
	docker-compose up --build -d
	@echo "✅ Приложение запущено! Доступно по адресу: http://localhost:8080/api/v1/efficiency?id=101"

# Остановка контейнеров
down:
	docker-compose down

# Просмотр логов контейнера приложения
logs:
	docker-compose logs -f app

# Перезапуск (остановить, собрать заново, запустить)
restart: down up

# --- Локальная разработка (без Docker-контейнера приложения) ---

# Запуск Go-приложения локально (БД должна быть запущена в Docker или локально)
run:
	go run cmd/app/main.go

# Запуск тестов
test:
	go test -v ./...

# Очистка мусора
clean:
	rm -f $(APP_NAME)
	docker system prune -f

# --- Работа с Базой Данных ---

# Подключение к БД внутри контейнера (открывает psql консоль)
db-shell:
	docker exec -it atm_db psql -U postgres -d atm_db

# Быстрое наполнение тестовыми данными (те, что мы писали выше)
seed:
	docker exec -i atm_db psql -U postgres -d atm_db < scripts/seed.sql
	@echo "✅ Тестовые данные загружены!"

# --- Миграции (если установлен golang-migrate) ---
# create-migration name=init_schema
migration-new:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down