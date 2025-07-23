package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adam-xu-mantle/go-template/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"

	// Database drivers
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	gorm *gorm.DB
}

// GetDatabaseDialector get database dialector from driver and dsn
func GetDatabaseDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "postgres", "postgresql":
		return postgres.Open(dsn), nil
	case "mysql":
		return mysql.Open(dsn), nil
	case "sqlite", "sqlite3":
		return sqlite.Open(dsn), nil
	case "sqlserver", "mssql":
		return sqlserver.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	// dsn := c.Database.Source
	// driver := c.Database.Driver

	// if dsn == "" {
	// 	return nil, func() {}, errors.New("database source is required")
	// }
	// if driver == "" {
	// 	return nil, func() {}, errors.New("database driver is required")
	// }

	// dialector, err := GetDatabaseDialector(driver, dsn)
	// if err != nil {
	// 	return nil, func() {}, fmt.Errorf("failed to get database dialector: %w", err)
	// }

	// gormConfig := gorm.Config{
	// 	Logger:                 NewLogger(logger),
	// 	SkipDefaultTransaction: true,
	// 	CreateBatchSize:        3_000,
	// }

	// gormdb, err := gorm.Open(dialector, &gormConfig)
	// if err != nil {
	// 	return nil, func() {}, fmt.Errorf("failed to connect to database (%s): %w", driver, err)
	// }

	// rawdb, _ := gormdb.DB()
	// rawdb.SetMaxOpenConns(200)
	// rawdb.SetConnMaxLifetime(time.Hour)
	// rawdb.SetConnMaxIdleTime(time.Minute * 5)
	// rawdb.SetMaxIdleConns(10)

	// db := &Data{
	// 	gorm: gormdb,
	// }

	// return db, func() {}, nil
	return nil, func() {}, nil
}

// GetDB return gorm db instance
func (d *Data) GetDB() *gorm.DB {
	return d.gorm
}

// ExecuteSQLMigration execute sql migration from migrationsFolder
func (db *Data) ExecuteSQLMigration(migrationsFolder string) error {
	err := filepath.Walk(migrationsFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to process migration file: %s", path))
		}
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".sql" {
			return nil
		}

		// #nosec G304 - path is validated above
		fileContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return errors.Wrap(readErr, fmt.Sprintf("Error reading SQL file: %s", path))
		}

		execErr := db.gorm.Exec(string(fileContent)).Error
		if execErr != nil {
			return errors.Wrap(execErr, fmt.Sprintf("Error executing SQL script: %s", path))
		}
		return nil
	})
	return err
}
