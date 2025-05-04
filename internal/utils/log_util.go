package utils

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Log утилита для логирования операций GORM
type Log struct {
	logger *logrus.Logger
}

// NewLog возвращает новый Log-объект
func NewLog(logger *logrus.Logger) *Log {
	return &Log{
		logger: logger,
	}
}

// AfterCreate hook срабатывает после создания
func (l *Log) AfterCreate(db *gorm.DB) {
	l.logger.WithFields(logrus.Fields{
		"table": db.Statement.Table,
		"event": "create",
		"data":  db.Statement.Dest,
	}).Info("GORM hook")
}

// AfterUpdate hook срабатывает после обновления
func (l *Log) AfterUpdate(db *gorm.DB) {
	l.logger.WithFields(logrus.Fields{
		"table": db.Statement.Table,
		"event": "update",
		"data":  db.Statement.Dest,
	}).Info("GORM hook")
}

// AfterDelete hook срабатывает после удаления
func (l *Log) AfterDelete(db *gorm.DB) {
	l.logger.WithFields(logrus.Fields{
		"table": db.Statement.Table,
		"event": "delete",
		"data":  db.Statement.Dest,
	}).Info("GORM hook")
}

// RegisterCallbacks регистрирует все hook-и
func (l *Log) RegisterCallbacks(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("log:after_create", l.AfterCreate)
	db.Callback().Update().After("gorm:update").Register("log:after_update", l.AfterUpdate)
	db.Callback().Delete().After("gorm:delete").Register("log:after_delete", l.AfterDelete)
}


import "github.com/gin-gonic/gin"

func LoggerMiddleware() gin.HandlerFunc {
	return gin.Logger() // Using Gin's default logger for simplicity
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery() // Using Gin's default recovery middleware
}
