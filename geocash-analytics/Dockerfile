# 👇 ИСПРАВЛЕНИЕ: Используем 'alpine' без цифр, чтобы взять самую последнюю версию (Latest)
FROM golang:alpine AS builder

WORKDIR /app

# Копируем файлы описания модулей
COPY go.mod go.sum ./

# Скачиваем библиотеки
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app/main.go

# 2. Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/server .

# Копируем конфиг
COPY --from=builder /app/config/config.yaml ./config/config.yaml

# Открываем порт
EXPOSE 8080

CMD ["./server"]