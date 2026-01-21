package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"geocash/internal/analytics" // Импорт интерфейсов и структур аналитики
	"geocash/internal/domain/terminal"
)

// AnalyticsRepository реализует интерфейс analytics.Repository
type AnalyticsRepository struct {
	db *sql.DB
}

// NewAnalyticsRepository создает новый экземпляр репозитория
func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{
		db: db,
	}
}

// GetPerformanceMetricsByPeriod выполняет сложный SQL-запрос для сбора метрик
func (r *AnalyticsRepository) GetPerformanceMetricsByPeriod(
	ctx context.Context,
	terminalID int,
	start time.Time,
	end time.Time,
) (analytics.PerformanceMetrics, error) {

	// SQL-запрос с использованием CTE для параллельного сбора статистики
	query := `
	WITH 
	-- 1. Считаем транзакции (количество и сумму)
	trans_stats AS (
		SELECT 
			COUNT(*) as total_count, 
			COALESCE(SUM(amount), 0) as total_amount
		FROM terminal_transactions
		WHERE terminal_id = $1 
		  AND transaction_time BETWEEN $2 AND $3
		  AND status = 'COMPLETED' -- Учитываем только успешные
	),
	-- 2. Считаем среднюю загрузку (в процентах: balance / capacity)
	load_stats AS (
		SELECT 
			COALESCE(AVG(current_balance / NULLIF(max_capacity, 0)), 0) as avg_load_percent
		FROM terminal_loadings
		WHERE terminal_id = $1 
		  AND record_time BETWEEN $2 AND $3
	),
	-- 3. Проверяем наличие критических ремонтов
	service_stats AS (
		SELECT EXISTS(
			SELECT 1 
			FROM terminal_service_history
			WHERE terminal_id = $1 
			  AND service_date BETWEEN $2 AND $3 
			  AND is_critical = true
		) as has_critical
	)
	-- Финальная выборка объединяет все результаты
	SELECT 
		t.total_count, 
		t.total_amount, 
		l.avg_load_percent, 
		s.has_critical
	FROM trans_stats t, load_stats l, service_stats s;
	`

	var metrics analytics.PerformanceMetrics

	// Выполняем запрос
	err := r.db.QueryRowContext(ctx, query, terminalID, start, end).Scan(
		&metrics.TotalTransactions,
		&metrics.TotalThroughputAmount,
		&metrics.AverageLoadingPercent,
		&metrics.LastServiceCriticality,
	)

	if err != nil {
		return analytics.PerformanceMetrics{}, fmt.Errorf("ошибка выполнения аналитического запроса: %w", err)
	}

	return metrics, nil
}

// GetLastKnownBalance получает последнюю запись о загрузке
func (r *AnalyticsRepository) GetLastKnownBalance(ctx context.Context, terminalID int) (terminal.CashBalance, error) {
	query := `
		SELECT record_time, current_balance, max_capacity
		FROM terminal_loadings
		WHERE terminal_id = $1
		ORDER BY record_time DESC
		LIMIT 1
	`

	var balance terminal.CashBalance
	balance.TerminalID = terminalID // ID мы уже знаем

	err := r.db.QueryRowContext(ctx, query, terminalID).Scan(
		&balance.RecordTime,
		&balance.CurrentBalance,
		&balance.MaxCapacity,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return terminal.CashBalance{}, fmt.Errorf("данные о балансе не найдены для терминала %d", terminalID)
		}
		return terminal.CashBalance{}, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	return balance, nil
}
