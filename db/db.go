package db

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const AppName = "astral-media"
const DefaultDatabaseName = "media.db"

const (
	EnvDatabasePath = "ASTRAL_MEDIA_DB_PATH"
	EnvConfigDir    = "ASTRAL_MEDIA_CONFIG_DIR"
)

type DB struct {
	*gorm.DB
}

func Open() (*DB, error) {
	path := FilePath()
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return &DB{DB: gormDB}, nil
}

func FilePath() string {
	if path := os.Getenv(EnvDatabasePath); path != "" {
		return path
	}

	if dir := os.Getenv(EnvConfigDir); dir != "" {
		return filepath.Join(dir, DefaultDatabaseName)
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultDatabaseName
	}

	return filepath.Join(cfgDir, "astrald", "apps", AppName, DefaultDatabaseName)
}
