package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	TLSCertFile string
	TLSKeyFile  string
}


func LoadConfig() *Config {
	
	
	
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "5000" 
	}

	return &Config{
		ServerPort:  port,
	}
}
