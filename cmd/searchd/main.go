package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"creatorsearch"
)

func main() {
	c, err := creatorsearch.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "q is required", http.StatusBadRequest)
			return
		}
		vec, err := c.Embed(r.Context(), query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		results, err := c.Query(context.Background(), "creator-content", vec, 5)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"query": query, "results": results, "count": strconv.Itoa(len(results))})
	})
	log.Println("searchd listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
