package config

import "os"

type Config struct {
	ServerPort string
}

func LoadConfig() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "5000" // Default port
	}
	return &Config{ServerPort: port}
}
