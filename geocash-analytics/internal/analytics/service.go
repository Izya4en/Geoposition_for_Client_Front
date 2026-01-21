package analytics

import (
	"math"
)

// GeoJSON структуры для вывода
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
}
type GeoJSONGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type GridService struct{}

func NewGridService() *GridService {
	return &GridService{}
}

// GenerateHexGrid создает сетку на весь город
func (s *GridService) GenerateHexGrid() GeoJSONFeatureCollection {
	// Координаты Астаны
	minLat, maxLat := 51.00, 51.30
	minLng, maxLng := 71.30, 71.65
	radius := 0.002

	var features []GeoJSONFeature
	h := radius * math.Sin(math.Pi/3)
	rowHeight := 1.5 * radius
	colWidth := 2 * h
	isOddRow := false

	for lat := minLat; lat < maxLat; lat += rowHeight {
		currLng := minLng
		if isOddRow {
			currLng += h
		}
		for lng := currLng; lng < maxLng; lng += colWidth {
			weight := s.calculateWeight(lat, lng)
			if weight > 0.05 {
				poly := s.createHexagon(lat, lng, radius)
				features = append(features, GeoJSONFeature{
					Type:       "Feature",
					Properties: map[string]interface{}{"weight": weight},
					Geometry:   GeoJSONGeometry{Type: "Polygon", Coordinates: [][][]float64{poly}},
				})
			}
		}
		isOddRow = !isOddRow
	}
	return GeoJSONFeatureCollection{Type: "FeatureCollection", Features: features}
}

func (s *GridService) calculateWeight(lat, lng float64) float64 {
	centerDist := math.Sqrt(math.Pow(lat-51.13, 2) + math.Pow(lng-71.43, 2))
	if centerDist > 0.16 {
		return 0
	}

	weight := 0.0
	// Шум
	noise := math.Sin(lat*400) * math.Cos(lng*400)
	weight += (noise + 1) * 0.15

	// Хотспоты
	hotspots := []struct{ lat, lng, p float64 }{
		{51.128, 71.430, 0.8}, {51.165, 71.425, 0.7}, {51.131, 71.402, 0.9},
	}
	for _, p := range hotspots {
		d := math.Sqrt(math.Pow(lat-p.lat, 2) + math.Pow(lng-p.lng, 2))
		if d < 0.035 {
			weight += p.p * (1.0 - d/0.035)
		}
	}
	if weight > 1.0 {
		return 1.0
	}
	return weight
}

func (s *GridService) createHexagon(lat, lng, r float64) [][]float64 {
	var coords [][]float64
	aspect := 1.65
	for i := 0; i <= 6; i++ {
		angle := math.Pi / 180 * (60.0*float64(i) - 30.0)
		coords = append(coords, []float64{lng + r*math.Cos(angle)*aspect, lat + r*math.Sin(angle)})
	}
	return coords
}
