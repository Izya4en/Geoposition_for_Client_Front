package main

import (
	"fmt"
	"net/http"

	"geocash/internal/analytics"
	"geocash/internal/dashboard"
	"geocash/internal/domain/terminal"
	"geocash/internal/platform/provider"
)

func main() {
	// 1. Инициализация зависимостей
	repo := terminal.NewMockRepository()
	gridSvc := analytics.NewGridService()
	osmProv := provider.NewOSMProvider()

	// 2. Инициализация Dashboard Service (Бизнес логика)
	dashSvc := dashboard.NewService(repo, osmProv, gridSvc)

	// 3. Инициализация Handler (HTTP слой)
	dashHandler := dashboard.NewHandler(dashSvc)

	// 4. Роутинг
	http.HandleFunc("/api/dashboard", dashHandler.ServeHTTP)

	// 5. Старт
	fmt.Println("🚀 GeoSmart Backend running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
