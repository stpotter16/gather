package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/stpotter16/gather/internal/handlers"
	"github.com/stpotter16/gather/internal/sessions"
	"github.com/stpotter16/gather/internal/store/postgres"
)

func run(
	ctx context.Context,
	getenv func(string) string,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL environment variable not set")
	}

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("initialising database: %w", err)
	}
	defer store.Close()

	sm, err := sessions.New(getenv("GATHER_HMAC_SECRET"), getenv("APP_ENV") == "production")
	if err != nil {
		return fmt.Errorf("initialising sessions: %w", err)
	}

	handler := handlers.NewServer(store, sm)

	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("Listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Received termination signal. Shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout, os.Stderr); err != nil {
		log.Fatalf("%s", err)
	}
}
