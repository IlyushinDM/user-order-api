package database

import (
	"errors"
	"testing"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/utils/config_util"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockLogger - простая оболочка для logrus.Logger
func mockLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

func validConfig() *config_util.Config {
	return &config_util.Config{
		DBHost:            "localhost",
		DBPort:            "5432",
		DBUser:            "user",
		DBName:            "testdb",
		DBPassword:        "password",
		DBMaxIdleConns:    1,
		DBMaxOpenConns:    2,
		DBConnMaxLifetime: time.Minute,
		DBConnMaxIdleTime: time.Second,
	}
}

// This test uses an invalid DSN to force a connection error.
func TestInitDB_InvalidDSN(t *testing.T) {
	cfg := validConfig()
	cfg.DBPort = "invalid" // Invalid port to force error
	log := mockLogger()
	db, err := InitDB(cfg, log)
	if db != nil {
		t.Error("expected db to be nil on connection error")
	}
	if err == nil {
		t.Error("expected error on invalid DSN")
	}
}

func TestRunMigrations_NilDB(t *testing.T) {
	log := mockLogger()
	err := RunMigrations(nil, log)
	if err == nil || err.Error() != "экземпляр *gorm.DB не предоставлен для выполнения миграций" {
		t.Errorf("непредвиденная ошибка: %v", err)
	}
}

func TestRunMigrations_NilLogger(t *testing.T) {
	// Use in-memory sqlite for test
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	err := RunMigrations(db, nil)
	if err == nil || err.Error() != "логгер не предоставлен для выполнения миграций" {
		t.Errorf("непредвиденная ошибка: %v", err)
	}
}

// This test checks that RunMigrations works with a valid DB and logger.
// It does not check actual DB schema, just that no error is returned.
func TestRunMigrations_Success(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	log := mockLogger()
	err := RunMigrations(db, log)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// Patch point for gorm.Open for testing
var gormOpen = func(dsn string, config *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), config)
}

// Optionally, test error propagation from gorm.Open
func TestInitDB_GormOpenError(t *testing.T) {
	cfg := validConfig()
	log := mockLogger()
	origOpen := gormOpen
	defer func() { gormOpen = origOpen }()
	gormOpen = func(dsn string, config *gorm.Config) (*gorm.DB, error) {
		return nil, errors.New("gorm open error")
	}
	db, err := InitDB(cfg, log)
	if db != nil {
		t.Error("expected db to be nil on gorm open error")
	}
	if err == nil || err.Error() == "" {
		t.Error("expected error from gorm open")
	}
}
