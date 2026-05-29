 package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/thesouldev/goboxd/internal/api"
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/queue"
)

var cfg *config.Config

func main() {
	var err error
	cfg, err = config.Load("config/languages.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("loaded %d languages\n", len(cfg.Languages))

	q := queue.New(runtime.NumCPU())
	api.Init(cfg, q)

	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/readyz", readyzHandler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/run", api.RunHandler)

	fmt.Println("goboxd starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	type LangStatus struct {
		OK      bool   `json:"ok"`
		Version string `json:"version,omitempty"`
		Error   string `json:"error,omitempty"`
	}
	type ReadyzResponse struct {
		Status    string                `json:"status"`
		Languages map[string]LangStatus `json:"languages"`
	}

	resp := ReadyzResponse{
		Status:    "ok",
		Languages: make(map[string]LangStatus),
	}

	// Dynamically check all languages from config
	// No hardcoding - adding a language to yaml automatically shows here
	for _, lang := range cfg.Languages {
		cmd := lang.Run.Cmd
		if lang.Build.Cmd != "" {
			cmd = lang.Build.Cmd
		}
		if _, err := os.Stat(cmd); err != nil {
			resp.Languages[lang.ID] = LangStatus{
				OK:    false,
				Error: cmd + " not found",
			}
			resp.Status = "degraded"
		} else {
			out, _ := exec.Command(cmd, "--version").CombinedOutput()
			version := string(out)
			if len(version) > 60 {
				version = version[:60]
			}
			resp.Languages[lang.ID] = LangStatus{OK: true, Version: version}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Status == "degraded" {
		w.WriteHeader(503)
	}
	json.NewEncoder(w).Encode(resp)
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	type LangInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type InfoResponse struct {
		BuildInfo map[string]string `json:"build_info"`
		Languages []LangInfo        `json:"languages"`
		Limits    map[string]int    `json:"limits"`
	}

	langs := make([]LangInfo, len(cfg.Languages))
	for i, l := range cfg.Languages {
		langs[i] = LangInfo{ID: l.ID, Name: l.Name}
	}

	resp := InfoResponse{
		BuildInfo: map[string]string{
			"version":    "0.1.0",
			"go_version": runtime.Version(),
		},
		Languages: langs,
		Limits: map[string]int{
			"max_source_bytes":    256 * 1024,
			"max_tests":           50,
			"max_concurrent_jobs": runtime.NumCPU(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}