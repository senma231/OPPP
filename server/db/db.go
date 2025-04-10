package db

import (
	"fmt"

	"github.com/senma231/p3/server/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// 设置日志级别
	logLevel := logger.Info
	if cfg.Log.Level == "debug" {
		logLevel = logger.Info
	} else if cfg.Log.Level == "error" {
		logLevel = logger.Error
	} else if cfg.Log.Level == "warn" {
		logLevel = logger.Warn
	} else if cfg.Log.Level == "silent" {
		logLevel = logger.Silent
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&User{},
		&Device{},
		&App{},
		&Forward{},
		&Connection{},
		&Stats{},
	); err != nil {
		return fmt.Errorf("自动迁移表结构失败: %w", err)
	}

	DB = db
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	return nil
}
