package traffic

// TrafficSegment - модель одной строки из CSV
type TrafficSegment struct {
	EdgeID         int64
	WeekdayTraffic int
	Geometry       string // WKT строка
}
