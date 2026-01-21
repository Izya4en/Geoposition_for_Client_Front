package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Драйвер для Postgres
	_ "github.com/lib/pq"

	"geocash/internal/analytics"
	"geocash/internal/dashboard"
	"geocash/internal/domain/terminal"
	"geocash/internal/platform/loader"
	"geocash/internal/platform/postgres"
	"geocash/internal/platform/provider"
)

func main() {
	// --- 1. ПОДКЛЮЧЕНИЕ К БАЗЕ ДАННЫХ ---
	// Берем настройки из переменных окружения или ставим дефолтные для localhost
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "atm_db")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	fmt.Println("🔌 Подключение к БД...", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка открытия соединения с БД: %v", err)
	}
	defer db.Close()

	// Проверяем пинг
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ БД недоступна: %v", err)
	}
	fmt.Println("✅ Успешное подключение к Postgres!")

	// --- 2. ИМПОРТ CSV (ТРАФИК) ---
	csvPath := "./traffic_data.csv"
	if _, err := os.Stat(csvPath); err == nil {
		fmt.Println("📂 Найден CSV файл, начинаем интеграцию...")

		// 1. Парсим CSV
		data, err := loader.LoadTrafficCSV(csvPath)
		if err != nil {
			log.Printf("❌ Ошибка чтения CSV: %v", err)
		} else {
			fmt.Printf("📊 Прочитано %d сегментов дорог.\n", len(data))

			// 2. Интегрируем в базу (обновляем зоны)
			integrator := postgres.NewTrafficIntegrator(db)

			// Используем таймаут для безопасности
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			err = integrator.EnrichZonesWithTraffic(ctx, data)
			if err != nil {
				log.Printf("❌ Ошибка интеграции в БД: %v", err)
			} else {
				fmt.Println("🚀 Успех! Зоны обновлены данными о трафике.")
				// Переименуем файл, чтобы не грузить его при каждом рестарте
				os.Rename(csvPath, csvPath+".processed")
			}
		}
	}

	// --- 3. ИНИЦИАЛИЗАЦИЯ СЕРВИСОВ ---

	// ВАЖНО: Сейчас здесь стоит Mock (фейковые данные).
	// Если у вас готов postgres.NewTerminalRepository(db), замените строку ниже на него.
	// Например: repo := postgres.NewTerminalRepository(db)
	repo := terminal.NewMockRepository()

	gridSvc := analytics.NewGridService()
	osmProv := provider.NewOSMProvider()

	// Инициализация Dashboard Service (Бизнес логика)
	dashSvc := dashboard.NewService(repo, osmProv, gridSvc)

	// Инициализация Handler (HTTP слой)
	dashHandler := dashboard.NewHandler(dashSvc)

	// --- 4. РОУТИНГ И СТАРТ ---
	http.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		// CORS заголовки для фронтенда
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		dashHandler.ServeHTTP(w, r)
	})

	fmt.Println("🚀 GeoSmart Backend running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// Вспомогательная функция для чтения ENV
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
