package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// initDatabase initializes the database connection and runs migrations
func initDatabase() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/relaxovision?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established")
	return db, nil
}

// runMigrations executes database migrations
func runMigrations(db *sql.DB) error {
	migrations := []string{
		"migrations/001_create_competitions.sql",
		"migrations/002_create_teams.sql",
		"migrations/003_create_matches.sql",
		"migrations/004_create_predictions.sql",
	}

	for _, migration := range migrations {
		slog.Info("Running migration", "file", migration)
		
		content, err := os.ReadFile(migration)
		if err != nil {
			// If migration files don't exist, skip them
			slog.Warn("Migration file not found, skipping", "file", migration)
			continue
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	slog.Info("All migrations completed successfully")
	return nil
}
