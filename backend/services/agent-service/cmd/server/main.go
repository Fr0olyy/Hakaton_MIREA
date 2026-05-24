package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dataforge/agent-service/internal/agent"
	"github.com/dataforge/agent-service/internal/config"
	"github.com/dataforge/agent-service/internal/handlers"
)

func main() {
	cfg := config.Load()

	agentSvc := agent.NewService(cfg.AgentBaseURL, cfg.AgentModel, cfg.AgentTimeout)
	handler := handlers.New(agentSvc)

	mux := http.NewServeMux()
	handler.Register(mux)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("agent-service listening on %s (ollama: %s, model: %s)",
			cfg.HTTPAddr, cfg.AgentBaseURL, cfg.AgentModel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}
