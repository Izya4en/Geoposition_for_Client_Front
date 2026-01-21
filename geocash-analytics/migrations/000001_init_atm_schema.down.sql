-- Удаляем таблицы в обратном порядке из-за внешних ключей (Foreign Keys)
DROP TABLE IF EXISTS efficiency_reports;
DROP TABLE IF EXISTS maintenance_logs;
DROP TABLE IF EXISTS cash_levels;
DROP TABLE IF EXISTS daily_stats;
DROP TABLE IF EXISTS terminals;

-- Опционально: отключаем PostGIS (если он больше не нужен другим таблицам)
-- DROP EXTENSION IF EXISTS postgis;