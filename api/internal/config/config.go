package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	ServerPort string
}

const envPath = ".env"

var current atomic.Pointer[Config]

// Load reads the .env file (if present) into the process environment and
// rebuilds the in-memory Config. Safe to call even when no DB is configured
// yet — the server boots in "unconfigured" mode until /setup writes one.
func Load() *Config {
	_ = godotenv.Load(envPath)

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	current.Store(cfg)
	return cfg
}

func Get() *Config {
	cfg := current.Load()
	if cfg == nil {
		return Load()
	}
	return cfg
}

func (c *Config) IsDBConfigured() bool {
	return c.DBHost != "" && c.DBUser != "" && c.DBName != ""
}

func MigrationsPath() string {
	p := os.Getenv("MIGRATIONS_PATH")
	if p == "" {
		p = "./migrations"
	}
	return p
}

type DBParams struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// Save persists DB connection params to .env (generating a JWT secret on
// first run if one doesn't exist yet), then reloads the in-memory Config so
// the running process picks it up without a restart.
func Save(params DBParams) (*Config, error) {
	existing, _ := godotenv.Read(envPath)
	if existing == nil {
		existing = map[string]string{}
	}

	existing["DB_HOST"] = params.Host
	existing["DB_PORT"] = params.Port
	existing["DB_USER"] = params.User
	existing["DB_PASSWORD"] = params.Password
	existing["DB_NAME"] = params.Name

	if existing["JWT_SECRET"] == "" {
		secret, err := generateSecret(32)
		if err != nil {
			return nil, err
		}
		existing["JWT_SECRET"] = secret
	}
	if existing["SERVER_PORT"] == "" {
		existing["SERVER_PORT"] = "8080"
	}

	if err := godotenv.Write(existing, envPath); err != nil {
		return nil, err
	}

	for k, v := range existing {
		os.Setenv(k, v)
	}

	return Load(), nil
}

func generateSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
