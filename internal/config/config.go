package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DB_SSLMODE         string
	DB_CHANNEL_BINDING string
	DBURL              string
}

func LoadConfig() Config {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")
	binding := os.Getenv("DB_CHANNEL_BINDING")

	// Ensure SSL is used
	if sslMode == "" {
		sslMode = "require"
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&channel_binding=%s",
		user,
		password,
		host,
		port,
		dbName,
		sslMode,
		binding,
	)

	return Config{
		DBHost:             host,
		DBPort:             port,
		DBUser:             user,
		DBPassword:         password,
		DBName:             dbName,
		DB_SSLMODE:         sslMode,
		DB_CHANNEL_BINDING: binding,
		DBURL:              dbURL,
	}
}
