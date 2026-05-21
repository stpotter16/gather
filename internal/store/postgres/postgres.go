package postgres

import (
	"cmp"
	"context"
	"embed"
	"fmt"
	"log"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

//go:embed migrations/*.sql
var migrationFS embed.FS

func New(ctx context.Context, databaseURL string) (Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return Store{}, fmt.Errorf("parsing database URL: %w", err)
	}
	// Required when connecting through PgBouncer in transaction mode (Neon pooler).
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return Store{}, fmt.Errorf("creating connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return Store{}, fmt.Errorf("pinging database: %w", err)
	}

	return Store{pool: pool}, nil
}

func (s Store) Close() {
	s.pool.Close()
}

func (s Store) RunMigrations(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	var currentVersion int
	row := s.pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	if err := row.Scan(&currentVersion); err != nil {
		return fmt.Errorf("reading current schema version: %w", err)
	}

	log.Printf("Database schema version: %d — %d migrations available", currentVersion, len(migrations))

	for _, m := range migrations {
		if currentVersion >= m.version {
			continue
		}

		migCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		tx, err := s.pool.Begin(migCtx)
		if err != nil {
			return fmt.Errorf("beginning migration %d transaction: %w", m.version, err)
		}

		if _, err := tx.Exec(migCtx, m.sql); err != nil {
			tx.Rollback(migCtx)
			return fmt.Errorf("executing migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(migCtx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version); err != nil {
			tx.Rollback(migCtx)
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}

		if err := tx.Commit(migCtx); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}

		log.Printf("Applied migration %d", m.version)
	}

	return nil
}

type migration struct {
	version int
	sql     string
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version, err := strconv.Atoi(entry.Name()[:3])
		if err != nil {
			return nil, fmt.Errorf("parsing version from %q: %w", entry.Name(), err)
		}

		sql, err := migrationFS.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration{version, string(sql)})
	}

	slices.SortFunc(migrations, func(a, b migration) int {
		return cmp.Compare(a.version, b.version)
	})

	return migrations, nil
}
