package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/stpotter16/gather/internal/store/postgres"
)

func run(ctx context.Context) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	databaseURL := os.Getenv("DATABASE_DIRECT_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_DIRECT_URL environment variable not set")
	}

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer store.Close()

	if err := store.RunMigrations(ctx); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("%s", err)
	}
}
