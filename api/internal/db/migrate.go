package db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending golang-migrate SQL migrations from
// migrationsPath against sqlDB. Migrations are the schema source of truth —
// GORM models are kept manually in sync and used only for querying.
func RunMigrations(sqlDB *sql.DB, migrationsPath string) error {
	driver, err := mysqlmigrate.WithInstance(sqlDB, &mysqlmigrate.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "mysql", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
