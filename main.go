package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	fmt.Println("server starting...")
	api := NewAPI()
	http.HandleFunc("/api", api.Handler)

	http.ListenAndServe(":8080", nil)
	fmt.Println("server started...")

}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in the handler")
	q := r.URL.Query().Get("q")
	data,CacheHit, err := a.getData(r.Context(), q)
	if err != nil {
		fmt.Printf("error getting data %v", err)
	}
	resp := APIResponse{
		Cache: CacheHit,
		Data:  data,
	}

	err = json.NewEncoder(w).Encode(resp)

	if err != nil {
		fmt.Printf("error encoding response %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *API) getData(ctx context.Context, q string) ([]NominatimResponse, bool, error) {
	value, err := a.cache.Get(ctx, q).Result()

	if err == redis.Nil {
		escapedQ := url.PathEscape(q)
		address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escapedQ)

		resp, err := http.Get(address)
		if err != nil {
			return nil, false, err
		}

		data := make([]NominatimResponse, 0)
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, false, err
		}

		b, err := json.Marshal(data)

		if err != nil {
			return nil, false, err
		}

		err = a.cache.Set(ctx, q, bytes.NewBuffer(b).Bytes(), time.Second*15).Err()
		if err != nil {
			return nil, false, err
		}

		return data, false, nil
	} else if err != nil {
		fmt.Printf("error calling redis %v", err)
		return nil, false, err
	} else {
		data := make([]NominatimResponse, 0)
		err := json.Unmarshal(bytes.NewBufferString(value).Bytes(), &data)

		if err != nil {
			return nil, false, err
		}
		return data, true, nil
	}

}

type API struct {
	cache *redis.Client
}

func NewAPI() *API {
	redisAddress := fmt.Sprintf("%s:6379", os.Getenv("REDIS_URL"))
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	return &API{
		cache: rdb,
	}
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
