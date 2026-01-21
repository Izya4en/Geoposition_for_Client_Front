package analytics

// PerformanceMetrics - структура для сбора статистики по терминалу
type PerformanceMetrics struct {
	TotalTransactions      int     // Количество транзакций
	TotalThroughputAmount  int     // Общая сумма
	AverageLoadingPercent  float64 // Средняя загрузка в %
	LastServiceCriticality bool    // Были ли критические ремонты
}
