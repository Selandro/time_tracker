package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env:"ENV" env-default:"local"`
	Database   DatabaseConfig   `yaml:"database"` // Database содержит настройки базы данных.
	HTTPServer HTTPServerConfig `yaml:"http_server"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`     // Host адрес хоста базы данных.
	Port     int    `yaml:"port"`     // Port порт базы данных.
	User     string `yaml:"user"`     // User имя пользователя базы данных.
	Password string `yaml:"password"` // Password пароль пользователя базы данных.
	DBName   string `yaml:"dbname"`   // DBName имя базы данных.
	SSLMode  string `yaml:"sslmode"`  // SSLMode режим SSL подключения.
}

type HTTPServerConfig struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	//необходимо установить переменную окружения к файлу ./servis/cmd/config/local.yaml
	configPath := os.Getenv("CONFIG_PATH_USER")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
