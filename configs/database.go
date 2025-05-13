package configs

import (
    "fmt"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// DBConfig 数据库配置
type DBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
}

// NewDBConfig 创建数据库配置
func NewDBConfig() *DBConfig {
    return &DBConfig{
        Host:     getEnvOrDefault("DB_HOST", "localhost"),
        Port:     getEnvOrDefault("DB_PORT", "5432"),
        User:     getEnvOrDefault("DB_USER", "postgres"),
        Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
        DBName:   getEnvOrDefault("DB_NAME", "simple_dsp"),
    }
}

// DSN 获取数据库连接字符串
func (c *DBConfig) DSN() string {
    return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
        c.Host, c.User, c.Password, c.DBName, c.Port)
}

// Connect 连接数据库
func (c *DBConfig) Connect() (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("连接数据库失败: %v", err)
    }
    return db, nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
} 