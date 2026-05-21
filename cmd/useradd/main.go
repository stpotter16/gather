package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"

	"github.com/stpotter16/gather/internal/password"
	"github.com/stpotter16/gather/internal/store/postgres"
	"golang.org/x/term"
)

var avatarColors = []string{
	"#fbbf24", // amber-400
	"#38bdf8", // sky-400
	"#fb7185", // rose-400
	"#a78bfa", // violet-400
	"#34d399", // emerald-400
	"#fb923c", // orange-400
	"#2dd4bf", // teal-400
	"#f472b6", // pink-400
}

func run(ctx context.Context) error {
	email := flag.String("email", "", "user email (required)")
	name := flag.String("name", "", "display name (required)")
	color := flag.String("color", "", "avatar hex color (default: random)")
	flag.Parse()

	if *email == "" || *name == "" {
		return errors.New("both -email and -name are required")
	}

	avatarColor := *color
	if avatarColor == "" {
		avatarColor = avatarColors[rand.IntN(len(avatarColors))]
	}

	fmt.Fprint(os.Stderr, "Password: ")
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}
	if len(raw) == 0 {
		return errors.New("password cannot be empty")
	}

	hash, err := password.Hash(string(raw))
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL not set")
	}

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer store.Close()

	id, err := store.CreateUser(ctx, *name, *email, avatarColor, hash)
	if err != nil {
		return err
	}

	fmt.Printf("Created user %d: %s <%s>\n", id, *name, *email)
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("%s", err)
	}
}
