package provider

import (
	"encoding/json"
	"geocash/internal/domain/terminal"
	"io"
	"net/http"
	"strings"
	"time"
)

type OSMProvider struct {
	client *http.Client
}

func NewOSMProvider() *OSMProvider {
	return &OSMProvider{client: &http.Client{Timeout: 10 * time.Second}}
}

func (p *OSMProvider) FetchAllATMs() ([]terminal.ATM, error) {
	// Запрос к Overpass: дай мне все ATM и Банки в Астане
	query := `[out:json][timeout:25];
		(
			node["amenity"="atm"](51.05,71.30,51.25,71.55);
			node["amenity"="bank"](51.05,71.30,51.25,71.55);
		);
		out body;`

	url := "https://overpass-api.de/api/interpreter"
	req, _ := http.NewRequest("POST", url, strings.NewReader(query))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var osmData struct {
		Elements []struct {
			ID   int64             `json:"id"`
			Lat  float64           `json:"lat"`
			Lon  float64           `json:"lon"`
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}
	if err := json.Unmarshal(body, &osmData); err != nil {
		return nil, err
	}

	var atms []terminal.ATM
	for _, el := range osmData.Elements {
		bank := "Unknown"
		// Пытаемся определить банк по тегам
		if val, ok := el.Tags["brand"]; ok {
			bank = val
		} else if val, ok := el.Tags["operator"]; ok {
			bank = val
		} else if val, ok := el.Tags["name"]; ok {
			bank = val
		}

		name := bank + " ATM"
		if val, ok := el.Tags["name"]; ok {
			name = val
		}

		// Мы БОЛЬШЕ НЕ пропускаем Forte. Мы берем всех.
		atms = append(atms, terminal.ATM{
			ID: int(el.ID), Name: name, Lat: el.Lat, Lng: el.Lon, Bank: bank,
		})
	}
	return atms, nil
}
