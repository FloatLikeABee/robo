package mysql

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/formsx/backend/internal/config"
	"github.com/formsx/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	path := cfg.FormsXSQLitePath
	if path == "" {
		return nil, fmt.Errorf("FORMSX_SQLITE_PATH is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sqlite mkdir: %w", err)
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&models.Form{},
		&models.FormPage{},
		&models.Question{},
		&models.QuestionRule{},
		&models.GraphSyncOutbox{},
	); err != nil {
		return nil, err
	}
	return db, nil
}
