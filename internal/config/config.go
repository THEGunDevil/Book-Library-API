package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBURL      string
}

func LoadConfig() Config {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Build a Postgres URL with SSL enabled
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=require",
		user,
		password,
		host,
		port,
		dbName,
	)

	return Config{
		DBHost:     host,
		DBPort:     port,
		DBUser:     user,
		DBPassword: password,
		DBName:     dbName,
		DBURL:      dbURL,
	}
}

