package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Name  string
	Port  string
	Env   string
	Debug bool
}

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	Params       string
	RootPassword string
}

type JWTConfig struct {
	Secret               string
	AccessExpiryMinutes  int
	RefreshExpiryHours   int
}

type Config struct {
	App         AppConfig
	DB          DBConfig
	JWT         JWTConfig
	FrontendURL string
}

// insecureDefaultSecrets are placeholder values from .env.example that must
// never reach a production deployment.
var insecureDefaultSecrets = map[string]bool{
	"":                                        true,
	"change-this-to-a-random-string":          true,
	"change-this-to-a-random-string-at-least-32-bytes-long": true,
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
		log.Println("no .env file found, relying on process environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Name:  v.GetString("APP_NAME"),
			Port:  v.GetString("APP_PORT"),
			Env:   v.GetString("APP_ENV"),
			Debug: v.GetBool("APP_DEBUG"),
		},
		DB: DBConfig{
			Host:         v.GetString("DB_HOST"),
			Port:         v.GetString("DB_PORT"),
			User:         v.GetString("DB_USER"),
			Password:     v.GetString("DB_PASSWORD"),
			Name:         v.GetString("DB_NAME"),
			Params:       v.GetString("DB_PARAMS"),
			RootPassword: v.GetString("DB_ROOT_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:              v.GetString("JWT_SECRET"),
			AccessExpiryMinutes: v.GetInt("JWT_ACCESS_EXPIRY_MINUTES"),
			RefreshExpiryHours:  v.GetInt("JWT_REFRESH_EXPIRY_HOURS"),
		},
		FrontendURL: v.GetString("FRONTEND_URL"),
	}

	if err := cfg.guardProductionSecrets(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// guardProductionSecrets fails fast instead of booting a production
// deployment with a placeholder/weak JWT secret left over from
// .env.example.
func (c *Config) guardProductionSecrets() error {
	if c.App.Env != "production" {
		return nil
	}
	if insecureDefaultSecrets[c.JWT.Secret] || len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET is missing, a known default, or shorter than 32 bytes — refusing to start in production")
	}
	return nil
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", d.User, d.Password, d.Host, d.Port, d.Name, d.Params)
}
