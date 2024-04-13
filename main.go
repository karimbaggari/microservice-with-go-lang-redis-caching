package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	fmt.Println("server starting...")
	http.HandleFunc("/api", Handler)
	
	http.ListenAndServe(":8080", nil)
	fmt.Println("server started...")

}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in the handler")
	q := r.URL.Query().Get("q")
	data, err := getData(q)
	if err != nil {
		fmt.Printf("error getting data %v", err)
	}
	resp := APIResponse{
		Cache: false,
		Data:  data,
	}

	err = json.NewEncoder(w).Encode(resp)

	if err != nil {
		fmt.Printf("error encoding response %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getData(q string) ([]NominatimResponse, error) {
	escapedQ := url.PathEscape(q)
	address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escapedQ)

	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}

	data := make([]NominatimResponse, 0)
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type APIResponse struct {
	Cache bool                `json:"cache"`
	Data  []NominatimResponse `json:"data"`
}

type NominatimResponse struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	Icon        string   `json:"icon"`
}
