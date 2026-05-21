package db

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const AppName = "media"
const DefaultDatabaseName = "media.db"

const (
	EnvDatabasePath = "ASTRAL_MEDIA_DB_PATH"
	EnvConfigDir    = "ASTRAL_MEDIA_CONFIG_DIR"
	EnvDataDir      = "ASTRAL_MEDIA_DATA_DIR"
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

func ConfigDir() string {
	if dir := os.Getenv(EnvConfigDir); dir != "" {
		return dir
	}

	base, err := defaultBaseDir()
	if err != nil {
		return ""
	}

	return appDir(base)
}

func DataDir() string {
	if dir := os.Getenv(EnvDataDir); dir != "" {
		return dir
	}

	base, err := defaultDataBaseDir()
	if err != nil {
		return ""
	}

	return appDir(base)
}

func FilePath() string {
	if path := os.Getenv(EnvDatabasePath); path != "" {
		return path
	}

	dir := DataDir()
	if dir == "" {
		return DefaultDatabaseName
	}

	return filepath.Join(dir, DefaultDatabaseName)
}

func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share"), nil
}

func defaultBaseDir() (string, error) {
	return defaultBaseDirFor(runtime.GOOS)
}

func defaultDataBaseDir() (string, error) {
	return defaultDataBaseDirFor(runtime.GOOS)
}

func defaultBaseDirFor(goos string) (string, error) {
	switch goos {
	case "linux":
		if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
			return dir, nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".config"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "Library", "Application Support"), nil
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return dir, nil
		}

		return os.UserConfigDir()
	default:
		return os.UserConfigDir()
	}
}

func defaultDataBaseDirFor(goos string) (string, error) {
	switch goos {
	case "linux":
		return userDataDir()
	default:
		return defaultBaseDirFor(goos)
	}
}

func appDir(base string) string {
	return filepath.Join(base, "astral", "apps", AppName)
}
