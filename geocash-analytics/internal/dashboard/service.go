package dashboard

import (
	"fmt"
	"geocash/internal/analytics"
	"geocash/internal/domain/terminal"
	"geocash/internal/platform/provider"
	"strings"
)

type Service struct {
	repo terminal.Repository
	osm  *provider.OSMProvider
	grid *analytics.GridService

	// Кэши для скорости
	forteCache []terminal.ATM
	compCache  []terminal.ATM
}

func NewService(repo terminal.Repository, osm *provider.OSMProvider, grid *analytics.GridService) *Service {
	s := &Service{repo: repo, osm: osm, grid: grid}
	go s.refreshData() // Запускаем обновление при старте
	return s
}

func (s *Service) refreshData() {
	fmt.Println("🔄 Updating ATM data from OpenStreetMap...")

	// 1. Получаем ВСЕ банкоматы города
	allATMs, err := s.osm.FetchAllATMs()
	if err != nil {
		fmt.Println("❌ OSM Error:", err)
		return
	}

	var forte []terminal.ATM
	var others []terminal.ATM

	// 2. Сортируем: Forte vs Остальные
	for i := range allATMs {
		atm := allATMs[i]

		// Проверка: это Forte?
		name := strings.ToLower(atm.Bank) + strings.ToLower(atm.Name)
		if strings.Contains(name, "forte") {
			// Это наш банкомат! Но в OSM нет данных о кассетах.
			// Генерируем их через MockRepo
			s.repo.EnrichATM(&atm)
			forte = append(forte, atm)
		} else {
			// Это конкурент
			atm.IsForte = false
			others = append(others, atm)
		}
	}

	s.forteCache = forte
	s.compCache = others
	fmt.Printf("✅ Data Updated: %d Forte ATMs, %d Competitors\n", len(forte), len(others))
}

func (s *Service) GetDashboardData() DashboardResponse {
	// Если кэш пуст (OSM еще не ответил), генерируем фейки
	competitors := s.compCache
	if len(competitors) == 0 {
		competitors = s.repo.GenerateRandomCompetitors(300)
	}

	// Forte тоже берем из кэша (если там пусто, можно вернуть старый хардкод, но OSM обычно находит)
	forte := s.forteCache

	return DashboardResponse{
		Forte:       forte,
		Competitors: competitors,
		HeatmapGrid: s.grid.GenerateHexGrid(),
	}
}
