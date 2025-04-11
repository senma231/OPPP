package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// 检查默认值
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("默认服务器主机错误，期望 0.0.0.0，实际 %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("默认服务器端口错误，期望 8080，实际 %d", cfg.Server.Port)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("默认数据库驱动错误，期望 postgres，实际 %s", cfg.Database.Driver)
	}
	if cfg.JWT.Secret == "" {
		t.Error("默认 JWT 密钥不应为空")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("P3_SERVER_PORT", "9090")
	os.Setenv("P3_DB_HOST", "test-db-host")
	os.Setenv("P3_JWT_SECRET", "test-jwt-secret")
	defer func() {
		os.Unsetenv("P3_SERVER_PORT")
		os.Unsetenv("P3_DB_HOST")
		os.Unsetenv("P3_JWT_SECRET")
	}()

	// 创建默认配置
	cfg := DefaultConfig()

	// 从环境变量加载配置
	loadFromEnv(cfg)

	// 检查环境变量是否正确加载
	if cfg.Server.Port != 9090 {
		t.Errorf("从环境变量加载服务器端口错误，期望 9090，实际 %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "test-db-host" {
		t.Errorf("从环境变量加载数据库主机错误，期望 test-db-host，实际 %s", cfg.Database.Host)
	}
	if cfg.JWT.Secret != "test-jwt-secret" {
		t.Errorf("从环境变量加载 JWT 密钥错误，期望 test-jwt-secret，实际 %s", cfg.JWT.Secret)
	}
}

func TestValidateConfig(t *testing.T) {
	// 测试有效配置
	validCfg := DefaultConfig()
	if err := validateConfig(validCfg); err != nil {
		t.Errorf("验证有效配置失败: %v", err)
	}

	// 测试无效服务器端口
	invalidPortCfg := DefaultConfig()
	invalidPortCfg.Server.Port = 0
	if err := validateConfig(invalidPortCfg); err == nil {
		t.Error("应该检测到无效的服务器端口")
	}

	// 测试无效数据库驱动
	invalidDBDriverCfg := DefaultConfig()
	invalidDBDriverCfg.Database.Driver = ""
	if err := validateConfig(invalidDBDriverCfg); err == nil {
		t.Error("应该检测到无效的数据库驱动")
	}

	// 测试无效 JWT 密钥
	invalidJWTSecretCfg := DefaultConfig()
	invalidJWTSecretCfg.JWT.Secret = ""
	if err := validateConfig(invalidJWTSecretCfg); err == nil {
		t.Error("应该检测到无效的 JWT 密钥")
	}
}

func TestGetDSN(t *testing.T) {
	// 测试 PostgreSQL DSN
	pgCfg := DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "p3",
		SSLMode:  "disable",
	}
	expectedPgDSN := "host=localhost port=5432 user=postgres password=postgres dbname=p3 sslmode=disable"
	if dsn := pgCfg.GetDSN(); dsn != expectedPgDSN {
		t.Errorf("PostgreSQL DSN 错误，期望 %s，实际 %s", expectedPgDSN, dsn)
	}

	// 测试 MySQL DSN
	mysqlCfg := DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "root",
		DBName:   "p3",
	}
	expectedMysqlDSN := "root:root@tcp(localhost:3306)/p3?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn := mysqlCfg.GetDSN(); dsn != expectedMysqlDSN {
		t.Errorf("MySQL DSN 错误，期望 %s，实际 %s", expectedMysqlDSN, dsn)
	}

	// 测试 SQLite DSN
	sqliteCfg := DatabaseConfig{
		Driver: "sqlite3",
		DBName: "p3.db",
	}
	expectedSqliteDSN := "p3.db"
	if dsn := sqliteCfg.GetDSN(); dsn != expectedSqliteDSN {
		t.Errorf("SQLite DSN 错误，期望 %s，实际 %s", expectedSqliteDSN, dsn)
	}

	// 测试未知驱动
	unknownCfg := DatabaseConfig{
		Driver: "unknown",
	}
	if dsn := unknownCfg.GetDSN(); dsn != "" {
		t.Errorf("未知驱动 DSN 错误，期望空字符串，实际 %s", dsn)
	}
}
