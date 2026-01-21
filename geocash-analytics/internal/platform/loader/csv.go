package loader

import (
	"encoding/csv"
	"fmt"
	"geocash/internal/domain/traffic" // Убедись, что этот пакет есть (entity.go)
	"io"
	"os"
	"strconv"
)

// LoadTrafficCSV читает файл и возвращает массив структур
func LoadTrafficCSV(path string) ([]traffic.TrafficSegment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	// r.Comma = ';' // Если вдруг разделитель точка с запятой

	// Пропускаем заголовок
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	var segments []traffic.TrafficSegment

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Парсим ID из научной нотации (1.91E+16)
		edgeIDFloat, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			continue // Пропускаем битые строки
		}

		// Парсим трафик (берем 2-ю колонку - weekday_traffic)
		wd, _ := strconv.Atoi(record[1])

		segments = append(segments, traffic.TrafficSegment{
			EdgeID:         int64(edgeIDFloat),
			WeekdayTraffic: wd,
			Geometry:       record[5], // Колонка geometry (WKT)
		})
	}

	return segments, nil
}
