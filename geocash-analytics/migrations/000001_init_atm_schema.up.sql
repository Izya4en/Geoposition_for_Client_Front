
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Справочник терминалов
CREATE TABLE IF NOT EXISTS terminals (
    id SERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) UNIQUE NOT NULL,
    model VARCHAR(100), -- Аппаратный ID (напр. 2524524645252)
    address TEXT,
    city VARCHAR(50) DEFAULT 'Astana',
    location GEOMETRY(Point, 4326),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. Зоны трафика (из 2ГИС)
CREATE TABLE IF NOT EXISTS geo_traffic_zones (
    id SERIAL PRIMARY KEY,
    zone_name VARCHAR(100),
    area_polygon GEOMETRY(Polygon, 4326),
    traffic_score INT, -- 1-100
    avg_pedestrians_daily INT,
    source_data VARCHAR(50) DEFAULT '2GIS'
);

-- 3. Статистика (Финансы и Проходимость)
CREATE TABLE IF NOT EXISTS daily_stats (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    report_date DATE NOT NULL,
    total_withdrawal_amount NUMERIC(15, 2) DEFAULT 0,
    total_deposit_amount NUMERIC(15, 2) DEFAULT 0,
    transaction_count INT DEFAULT 0,
    unique_users_count INT DEFAULT 0,
    UNIQUE(terminal_id, report_date)
);

-- 4. Мониторинг наличности (Cash)
CREATE TABLE IF NOT EXISTS cash_levels (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    check_time TIMESTAMP NOT NULL,
    current_balance NUMERIC(15, 2) NOT NULL,
    max_capacity NUMERIC(15, 2) NOT NULL,
    load_percentage DECIMAL(5, 2),
    is_encashment_needed BOOLEAN DEFAULT FALSE
);

-- 5. Жалобы клиентов (Voice of Customer)
CREATE TABLE IF NOT EXISTS client_complaints (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE SET NULL,
    complaint_category VARCHAR(50), -- DIRTY, EATS_CARD, QUEUE
    complaint_text TEXT,
    status VARCHAR(20) DEFAULT 'OPEN',
    user_location GEOMETRY(Point, 4326),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Логи обслуживания
CREATE TABLE IF NOT EXISTS maintenance_logs (
    id SERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    issue_type VARCHAR(50),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    downtime_minutes INT
);

-- 7. Итоговые отчеты (Кэш для аналитики)
CREATE TABLE IF NOT EXISTS efficiency_reports (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    period_start DATE,
    period_end DATE,
    efficiency_status VARCHAR(20), -- EFFECTIVE / INEFFECTIVE
    recommendation TEXT,
    calculated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_terminals_geom ON terminals USING GIST (location);
CREATE INDEX idx_zones_geom ON geo_traffic_zones USING GIST (area_polygon);
CREATE INDEX idx_stats_date ON daily_stats (report_date);

-- VIEW 1: Главная сводка для карты (Heatmap)
-- Показывает: Где стоит, сколько денег приносит, есть ли активные жалобы.
CREATE OR REPLACE VIEW view_dashboard_map AS
SELECT 
    t.terminal_id,
    t.address,
    t.location,
    COALESCE(ds.total_withdrawal_amount, 0) as yesterday_turnover,
    COALESCE(ds.transaction_count, 0) as yesterday_traffic,
    (SELECT COUNT(*) FROM client_complaints c WHERE c.terminal_id = t.terminal_id AND c.status = 'OPEN') as active_complaints_count,
    CASE 
        WHEN ds.transaction_count < 5 THEN 'CRITICAL_LOW'
        WHEN (SELECT COUNT(*) FROM client_complaints c WHERE c.terminal_id = t.terminal_id AND c.status = 'OPEN') > 0 THEN 'HAS_ISSUES'
        ELSE 'NORMAL'
    END as map_status_color
FROM terminals t
LEFT JOIN daily_stats ds ON t.terminal_id = ds.terminal_id AND ds.report_date = (CURRENT_DATE - INTERVAL '1 day')::date;

-- VIEW 2: Анализ эффективности (GAP Analysis)
-- Показывает: Районы с высоким трафиком (2GIS), где у нас НЕТ банкоматов.
-- Логика: Берем зоны трафика и проверяем, попадает ли туда хоть один наш терминал.
CREATE OR REPLACE VIEW view_expansion_recommendations AS
SELECT 
    z.zone_name,
    z.traffic_score,
    z.avg_pedestrians_daily,
    z.area_polygon
FROM geo_traffic_zones z
WHERE NOT EXISTS (
    SELECT 1 
    FROM terminals t 
    WHERE ST_Contains(z.area_polygon, t.location)
)
AND z.traffic_score > 80; -- Только горячие зоны

-- 1. Терминалы
INSERT INTO terminals (terminal_id, model, address, location) VALUES
('AST-001', '2524524645252', 'Mega Silk Way', ST_SetSRID(ST_Point(71.4138, 51.0885), 4326)),
('AST-002', '8934001239481', 'Nuraly Zhol Station', ST_SetSRID(ST_Point(71.5464, 51.1243), 4326)),
('AST-088', '1111222233334', 'Industrial Zone A', ST_SetSRID(ST_Point(71.4900, 51.0500), 4326))
ON CONFLICT (terminal_id) DO NOTHING;

-- 2. Статистика
INSERT INTO daily_stats (terminal_id, report_date, total_withdrawal_amount, transaction_count) VALUES
('AST-001', CURRENT_DATE - INTERVAL '1 day', 15400000.00, 450),
('AST-002', CURRENT_DATE - INTERVAL '1 day', 9100000.00, 320),
('AST-088', CURRENT_DATE - INTERVAL '1 day', 50000.00, 3)
ON CONFLICT DO NOTHING;

-- 3. Зоны трафика 2GIS
INSERT INTO geo_traffic_zones (zone_name, traffic_score, avg_pedestrians_daily, area_polygon) VALUES
('EXPO District', 95, 15000, ST_GeomFromText('POLYGON((71.410 51.090, 71.420 51.090, 71.420 51.080, 71.410 51.080, 71.410 51.090))', 4326)),
('New Residential Area (Empty)', 85, 8000, ST_GeomFromText('POLYGON((71.450 51.150, 71.460 51.150, 71.460 51.140, 71.450 51.140, 71.450 51.150))', 4326));

-- 4. Жалобы
INSERT INTO client_complaints (terminal_id, complaint_category, status, user_location) VALUES
('AST-002', 'DIRTY', 'OPEN', ST_SetSRID(ST_Point(71.5465, 51.1244), 4326));