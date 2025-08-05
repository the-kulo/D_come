package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port int
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_IDLE"))
	maxOpen, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_OPEN"))
	maxLifetime, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_LIFETIME"))
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	return &Config{
		Database: DatabaseConfig{
			Host:        os.Getenv("MYSQL_HOST"),
			Port:        os.Getenv("MYSQL_PORT"),
			User:        os.Getenv("MYSQL_USER"),
			Password:    os.Getenv("MYSQL_PASSWORD"),
			DBName:      os.Getenv("MYSQL_DATABASE"),
			MaxIdle:     maxIdle,
			MaxOpen:     maxOpen,
			MaxLifetime: time.Duration(maxLifetime) * time.Second,
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
	}, nil
}
