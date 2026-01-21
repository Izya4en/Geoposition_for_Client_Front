package dashboard

import (
	"geocash/internal/analytics"
	"geocash/internal/domain/terminal"
)

type DashboardResponse struct {
	Forte       []terminal.ATM                     `json:"forte"`
	Competitors []terminal.ATM                     `json:"competitors"`
	HeatmapGrid analytics.GeoJSONFeatureCollection `json:"heatmapGrid"`
}
