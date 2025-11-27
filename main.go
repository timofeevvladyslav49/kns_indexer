package main

import (
	"context"
	_ "kns-indexer/docs"
	"kns-indexer/handlers"
	"kns-indexer/indexer"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/swagger/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title KNS Indexer API
// @version 1.0
// @description This is a simple API for KNS Indexer
// @BasePath /
func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}

	createTablesSql := `
	CREATE TABLE IF NOT EXISTS settings(
		page INTEGER NOT NULL CHECK (page > 0) DEFAULT 1,
		last_block_timestamp TIMESTAMPTZ,
		last_block_hash TEXT
	);
	CREATE TABLE IF NOT EXISTS username(
		username TEXT PRIMARY KEY,
		address TEXT NOT NULL,
		owner TEXT NOT NULL,
		cid TEXT,
		is_primary BOOLEAN NOT NULL DEFAULT FALSE,
		timestamp TIMESTAMPTZ NOT NULL
	);
	`

	if _, err = conn.Exec(context.Background(), createTablesSql); err != nil {
		panic(err)
	}

	var isSettingsExists bool
	if err = conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM settings);").Scan(&isSettingsExists); err != nil {
		panic(err)
	}
	if !isSettingsExists {
		if _, err = conn.Exec(context.Background(), "INSERT INTO settings DEFAULT VALUES;"); err != nil {
			panic(err)
		}
		slog.Info("Settings created")
	}

	conn.Release()

	slog.Info("Starting KNS Indexer")

	go indexer.Run(pool)

	app := fiber.New()
	app.Use(logger.New())

	app.Get("/*", handlers.NewDomainHandler(pool))

	app.Get("/docs/*", swagger.HandlerDefault)

	app.Get("/usernames", handlers.NewGetUsernamesHandler(pool))
	app.Get("/usernames/owner/:owner", handlers.NewGetOwnerUsernamesHandler(pool))
	app.Get("/usernames/:username", handlers.NewGetUsernameHandler(pool))
	app.Get("/primary-username/:owner", handlers.NewGetPrimaryUsernameHandler(pool))

	slog.Info("Starting API on :8000")

	if err = app.Listen(":8000"); err != nil {
		panic(err)
	}

	slog.Info("KNS Indexer stopped!")
}
