package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ayechanhan/user-records-app/backend/migrations"
)

// Connect opens a GORM connection to Postgres using the given DSN.
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}
	return db, nil
}

// Migrate applies all embedded SQL migrations to the database at dsn.
func Migrate(dsn string) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("postgres: migration source: %w", err)
	}

	sqlDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("postgres: connect for migration: %w", err)
	}
	genericDB, err := sqlDB.DB()
	if err != nil {
		return fmt.Errorf("postgres: extract sql.DB: %w", err)
	}
	defer genericDB.Close()

	dbDriver, err := pgmigrate.WithInstance(genericDB, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("postgres: migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("postgres: init migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("postgres: apply migrations: %w", err)
	}
	return nil
}
