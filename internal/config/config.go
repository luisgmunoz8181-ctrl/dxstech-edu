package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv  string
	Port    string
	Host    string
	DataDir string
}

func Load() *Config {
	env := os.Getenv("APP_ENV")
	if strings.TrimSpace(env) == "" {
		env = "development"
	}

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "3000"
	}

	host := os.Getenv("HOST")
	if strings.TrimSpace(host) == "" {
		host = "0.0.0.0"
	}

	dataDir := os.Getenv("DATA_DIR")
	if strings.TrimSpace(dataDir) == "" {
		dataDir = "./data"
	}

	return &Config{
		AppEnv:  env,
		Port:    port,
		Host:    host,
		DataDir: dataDir,
	}
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.AppEnv) == "development"
}

func (c *Config) PublicURL() string {
	return "http://localhost:" + c.Port
}

func (c *Config) BindAddress() string {
	return c.Host + ":" + c.Port
}
