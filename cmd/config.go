package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Host            string
	JWTKey          []byte
	GRPCServerAddr  string
	GRPCGatewayAddr string
	GinServerAddr   string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "testSecret"
		log.Println("JWT_SECRET should be set in an .env file. Using the default is not secure!")
	}

	return &Config{
		Host:            host,
		JWTKey:          []byte(jwtSecret),
		GRPCServerAddr:  host + ":9090",
		GRPCGatewayAddr: ":" + "8081",
		GinServerAddr:   ":" + "8080",
	}, nil
}
