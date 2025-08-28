package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GinMode         string
	Host            string
	GRPCServerPort  string
	GRPCGatewayPort string
	GinServerPort   string
	DB_DSN          string
	JWTSecret       string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LoadConfig checks for the existence of an .env file. If it doens't exist or if the variables are not sets,
// it sets the Config struct with default values.
//
// I know that the JWT secret may be always set, but using a default value and warning the user of the program
// can be a reasonable trade-off.
func LoadConfig() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	cfg := Config{}
	ginMode := getEnv("GIN_MODE", "release")
	if ginMode != "debug" && ginMode != "release" {
		return nil, fmt.Errorf("invalid GIN_MODE: %s, must be 'debug' or 'release'", ginMode)
	}
	cfg.GinMode = ginMode

	cfg.Host = getEnv("HOST", "localhost")
	cfg.GRPCServerPort = getEnv("GRPC_SERVER_PORT", "9090")
	cfg.GRPCGatewayPort = getEnv("GRPC_GATEWAY_PORT", "8081")
	cfg.GinServerPort = getEnv("GIN_SERVER_PORT", "8080")
	cfg.DB_DSN = getEnv("DB_DSN", "test.db")
	jwtSecret := getEnv("JWT_SECRET", "default_secret")
	if jwtSecret == "default_secret" {
		log.Println("WARNING: default vale for JWT_SECRET. In production, you must set a custom value.")
	}
	cfg.JWTSecret = jwtSecret

	log.Printf("Configuration loaded for %s environment", cfg.GinMode)
	return &cfg, nil
}
