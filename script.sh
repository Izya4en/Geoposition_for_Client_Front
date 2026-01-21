#!/bin/bash

# 1. Название проекта и модуля
PROJECT_NAME="geocash-analytics"
MODULE_NAME="geocash"

echo "🚀 Создание проекта $PROJECT_NAME..."

# Создаем корневую директорию и заходим в нее
mkdir -p $PROJECT_NAME
cd $PROJECT_NAME

# 2. Инициализация Go модуля
if command -v go &> /dev/null; then
    go mod init $MODULE_NAME
    echo "✅ Go модуль инициализирован"
else
    echo "⚠️ Go не найден, создаю пустой go.mod"
    touch go.mod
fi

# 3. Создание структуры папок (Standard Go Layout)
echo "📂 Создание структуры папок..."

# Cmd (Точки входа)
mkdir -p cmd/api
mkdir -p cmd/worker

# Config
mkdir -p config

# Internal (Основной код)
# Domain (Бизнес-логика)
mkdir -p internal/domain/terminal
mkdir -p internal/domain/monitoring
mkdir -p internal/domain/cash

# Services (Аналитика и Дашборд)
mkdir -p internal/analytics
mkdir -p internal/dashboard

# Platform (Инфраструктура и БД)
mkdir -p internal/platform/postgres
mkdir -p internal/platform/provider/twogis

# Pkg (Общие библиотеки)
mkdir -p pkg/logger
mkdir -p pkg/validator

# Migrations (SQL)
mkdir -p migrations

# 4. Создание базовых файлов с package именами

# --- CMD ---
cat <<EOF > cmd/api/main.go
package main

import (
	"fmt"
	"$MODULE_NAME/internal/dashboard"
)

func main() {
	fmt.Println("Starting GeoCash Analytics API...")
	// Здесь будет инициализация (DI)
}
EOF

cat <<EOF > cmd/worker/main.go
package main

import "fmt"

func main() {
	fmt.Println("Starting Background Worker...")
}
EOF

# --- DOMAIN (Terminal) ---
cat <<EOF > internal/domain/terminal/entity.go
package terminal

// Terminal - сущность банкомата
type Terminal struct {
	ID       string
	Location Location
	Address  string
}

type Location struct {
	Lat float64
	Lon float64
}
EOF

cat <<EOF > internal/domain/terminal/repository.go
package terminal

import "context"

type Repository interface {
	GetByID(ctx context.Context, id string) (*Terminal, error)
	GetAll(ctx context.Context) ([]Terminal, error)
}
EOF

# --- DOMAIN (Monitoring) ---
cat <<EOF > internal/domain/monitoring/entity.go
package monitoring

type Status string

const (
	StatusOnline  Status = "ONLINE"
	StatusOffline Status = "OFFLINE"
)
EOF

# --- ANALYTICS ---
cat <<EOF > internal/analytics/service.go
package analytics

import "context"

type Service struct {
	// repo Repository
}

func (s *Service) GetForecast(ctx context.Context, terminalID string) (float64, error) {
	return 0.0, nil
}
EOF

cat <<EOF > internal/analytics/repository.go
package analytics

// Здесь будут сложные SQL запросы (PostGIS)
type Repository interface {
	GetClusterData(cityID string)
}
EOF

# --- DASHBOARD (BFF) ---
cat <<EOF > internal/dashboard/service.go
package dashboard

import (
	"$MODULE_NAME/internal/domain/terminal"
)

// Service оркестрирует получение данных
type Service struct {
	termRepo terminal.Repository
}
EOF

cat <<EOF > internal/dashboard/dto.go
package dashboard

type MapPointDTO struct {
	ID    string  \`json:"id"\`
	Lat   float64 \`json:"lat"\`
	Lon   float64 \`json:"lon"\`
	Color string  \`json:"color"\`
}
EOF

cat <<EOF > internal/dashboard/handler.go
package dashboard

import "net/http"

type Handler struct {
	svc *Service
}

func (h *Handler) GetMap(w http.ResponseWriter, r *http.Request) {
	// Call service
}
EOF

# --- PLATFORM ---
cat <<EOF > internal/platform/postgres/connection.go
package postgres

// InitDB connection logic
func Connect(dsn string) {
	// pgxpool.Connect...
}
EOF

# --- CONFIG & FILES ---
touch config/config.yaml
touch .gitignore
touch Makefile

# Dockerfile
cat <<EOF > Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
EOF

echo "✅ Проект $PROJECT_NAME успешно создан!"
echo "👉 cd $PROJECT_NAME"