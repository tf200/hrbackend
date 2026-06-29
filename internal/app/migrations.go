package app

import (
	"context"
	"errors"
	"fmt"

	"hrbackend/config"
	"hrbackend/internal/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(ctx context.Context, cfg config.Config, logger domain.Logger) error {
	migrator, err := migrate.New(cfg.MigrationsPath, cfg.DbSource)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		_, _ = migrator.Close()
	}()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.LogInfo(ctx, "migrations", "database migrations are up to date")
			return nil
		}

		return fmt.Errorf("run migrations: %w", err)
	}

	logger.LogInfo(ctx, "migrations", "database migrations applied")
	return nil
}
