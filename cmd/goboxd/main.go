package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/thesouldev/goboxd/internal/api"
)

func main() {
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/run", api.RunHandler)

	fmt.Println("goboxd starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}