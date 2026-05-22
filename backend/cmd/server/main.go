package main

import (
	"log"
	"net/http"

	"llm-gateway/backend/internal/config"
	"llm-gateway/backend/internal/database"
	httpapi "llm-gateway/backend/internal/http"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	router := httpapi.NewRouter(cfg, db)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	log.Printf("llm-gateway backend listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
