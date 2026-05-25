package main

import (
	"encoding/json"
	"net/http"
)

// handleGetCities возвращает список всех городов
func handleGetCities(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, name FROM cities ORDER BY name ASC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var city City
		if err := rows.Scan(&city.ID, &city.Name); err == nil {
			cities = append(cities, city)
		}
	}

	if cities == nil {
		cities = []City{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}
