package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPPort          string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	TxMaxRetries      int
	ShutdownTimeout   time.Duration
}

func Load(path string) (cfg Config, err error) {
	values, err := readFile(path)
	if err != nil {
		return cfg, err
	}
	cfg.HTTPPort = values["APP_PORT"]
	cfg.DBHost = values["DB_HOST"]
	cfg.DBPort = values["DB_PORT"]
	cfg.DBUser = values["DB_USER"]
	cfg.DBPassword = values["DB_PASSWORD"]
	cfg.DBName = values["DB_NAME"]
	cfg.DBSSLMode = values["DB_SSLMODE"]
	cfg.DBMaxOpenConns, err = parseInt(values["DB_MAX_OPEN_CONNS"], 50)
	if err != nil {
		return cfg, err
	}
	cfg.DBMaxIdleConns, err = parseInt(values["DB_MAX_IDLE_CONNS"], 25)
	if err != nil {
		return cfg, err
	}
	cfg.DBConnMaxLifetime, err = parseDuration(values["DB_CONN_MAX_LIFETIME"], "30m")
	if err != nil {
		return cfg, err
	}
	cfg.TxMaxRetries, err = parseInt(values["TX_MAX_RETRIES"], 5)
	if err != nil {
		return cfg, err
	}
	cfg.ShutdownTimeout, err = parseDuration(values["SHUTDOWN_TIMEOUT"], "10s")
	if err != nil {
		return cfg, err
	}
	return cfg, cfg.validate()
}

func (c Config) validate() error {
	if c.HTTPPort == "" || c.DBHost == "" || c.DBPort == "" || c.DBUser == "" || c.DBPassword == "" || c.DBName == "" || c.DBSSLMode == "" {
		return errors.New("missing required config value")
	}
	return nil
}

func readFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid config line")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func parseInt(value string, def int) (int, error) {
	if value == "" {
		return def, nil
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func parseDuration(value string, def string) (time.Duration, error) {
	if value == "" {
		value = def
	}
	return time.ParseDuration(value)
}
