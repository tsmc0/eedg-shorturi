package config

import (
	"os"
	"strconv"
	"time"
)

// Config — конфигурация сервиса.
type Config struct {
	ListenAddr    string
	DBPath        string
	BaseShortURL  string
	DefaultTTL    time.Duration
	MaxContentMB  int64
	APIKey        string
	RateLimitRPM  int
	CleanupPeriod time.Duration
	FetchTimeout  time.Duration
}

// Load читает конфиг из переменных окружения с разумными дефолтами.
func Load() Config {
	return Config{
		ListenAddr:    env("LISTEN_ADDR", "127.0.0.1:5443"),
		DBPath:        env("DB_PATH", "./data/shortlink.db"),
		BaseShortURL:  env("BASE_SHORT_URL", "https://s.eedg.pw"),
		DefaultTTL:    envDuration("DEFAULT_TTL", 1*time.Hour),
		MaxContentMB:  envInt64("MAX_CONTENT_MB", 5),
		APIKey:        env("API_KEY", "hegxv5#:>I\\l`eHbk%TsI3gB|%8LP3sl+172d2d\"(UypMNkzUs"), // пусто = авторизация выключена
		RateLimitRPM:  envInt("RATE_LIMIT_RPM", 120),
		CleanupPeriod: envDuration("CLEANUP_PERIOD", 10*time.Minute),
		FetchTimeout:  envDuration("FETCH_TIMEOUT", 15*time.Second),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envInt64(k string, def int64) int64 {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
