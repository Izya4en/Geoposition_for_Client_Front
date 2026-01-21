# 1. Сборка
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Копируем всё содержимое проекта (включая созданную папку vendor)
COPY . .

# Собираем приложение
# Флаг -mod=vendor заставляет Go брать библиотеки из папки, а не из интернета
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o server ./cmd/app

# 2. Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/server .

# Копируем конфиг (обязательно сохраняем структуру папок)
COPY --from=builder /app/config/config.yaml ./config/config.yaml

# Открываем порт
EXPOSE 8080

CMD ["./server"]