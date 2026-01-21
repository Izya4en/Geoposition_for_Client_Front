package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"geocash/internal/domain/traffic"
	"log"
	"strings"
)

type TrafficIntegrator struct {
	db *sql.DB
}

func NewTrafficIntegrator(db *sql.DB) *TrafficIntegrator {
	return &TrafficIntegrator{db: db}
}

// EnrichZonesWithTraffic обновляет traffic_score в таблице geo_traffic_zones
func (t *TrafficIntegrator) EnrichZonesWithTraffic(ctx context.Context, segments []traffic.TrafficSegment) error {
	if len(segments) == 0 {
		return nil
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Создаем временную таблицу
	_, err = tx.ExecContext(ctx, `CREATE TEMP TABLE temp_csv_traffic (traffic INT, geom TEXT) ON COMMIT DROP;`)
	if err != nil {
		return fmt.Errorf("ошибка создания temp таблицы: %w", err)
	}

	// 2. Заливаем данные (Batch Insert)
	// Берем первые 2000 для примера, чтобы не забить память
	limit := 2000
	if len(segments) < limit {
		limit = len(segments)
	}

	valueStrings := make([]string, 0, limit)
	valueArgs := make([]interface{}, 0, limit*2)

	for i := 0; i < limit; i++ {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, segments[i].WeekdayTraffic, segments[i].Geometry)
	}

	stmt := fmt.Sprintf("INSERT INTO temp_csv_traffic (traffic, geom) VALUES %s", strings.Join(valueStrings, ","))
	_, err = tx.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("ошибка вставки batch: %w", err)
	}

	// 3. Обновляем зоны через пересечение (ST_Intersects)
	query := `
		UPDATE geo_traffic_zones z
		SET traffic_score = sub.total_traffic / 100
		FROM (
			SELECT z.id, SUM(t.traffic) as total_traffic
			FROM geo_traffic_zones z
			JOIN temp_csv_traffic t 
			ON ST_Intersects(z.area_polygon, ST_GeomFromText(t.geom, 4326))
			GROUP BY z.id
		) sub
		WHERE z.id = sub.id;
	`
	res, err := tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка update: %w", err)
	}

	count, _ := res.RowsAffected()
	log.Printf("✅ Обновлено зон трафика: %d", count)

	return tx.Commit()
}
