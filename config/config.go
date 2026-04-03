package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DbSource              string        `mapstructure:"DB_SOURCE"`
	ServerAddress         string        `mapstructure:"SERVER_ADDRESS"`
	AccessTokenSecretKey  string        `mapstructure:"ACCESS_TOKEN_SECRET_KEY"`
	AccessTokenDuration   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenSecretKey string        `mapstructure:"REFRESH_TOKEN_SECRET_KEY"`
	RefreshTokenDuration  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	TwoFATokenSecretKey   string        `mapstructure:"TWO_FA_TOKEN_SECRET_KEY"`
	TwoFATokenDuration    time.Duration `mapstructure:"TWO_FA_TOKEN_DURATION"`
	B2Endpoint            string        `mapstructure:"B2_ENDPOINT"`
	B2Key                 string        `mapstructure:"B2_KEY"`
	B2KeyID               string        `mapstructure:"B2_KEY_ID"`
	B2Bucket              string        `mapstructure:"B2_BUCKET"`
	DisableBucket         bool          `mapstructure:"DISABLE_BUCKET"`
	Host                  string        `mapstructure:"HOST"`
	RedisHost             string        `mapstructure:"REDIS_HOST"`
	RedisPassword         string        `mapstructure:"REDIS_PASSWORD"`
	Remote                bool          `mapstructure:"REMOTE"`
	OpenRouterAPIKey      string        `mapstructure:"OPEN_ROUTER_API_KEY"`
	SmtpName              string        `mapstructure:"SMTP_NAME"`
	SmtpAddress           string        `mapstructure:"SMTP_ADDRESS"`
	SmtpAuth              string        `mapstructure:"SMTP_AUTH"`
	SmtpHost              string        `mapstructure:"SMTP_HOST"`
	SmtpPort              int           `mapstructure:"SMTP_PORT"`
	BrevoSenderName       string        `mapstructure:"BREVO_SENDER_NAME"`
	BrevoSenderEmail      string        `mapstructure:"BREVO_SENDER_EMAIL"`
	BrevoApiKey           string        `mapstructure:"BREVO_API_KEY"`
	Environment           string        `mapstructure:"ENVIRONMENT"`
	OpenRouterApiKey      string        `mapstructure:"OPEN_ROUTER_API_KEY"`
	OpenRouterModel       string        `mapstructure:"OPEN_ROUTER_MODEL"`
	MigrationsPath        string        `mapstructure:"MIGRATIONS_PATH"`
	AdminEmail            string        `mapstructure:"ADMIN_EMAIL"`
	AdminPassword         string        `mapstructure:"ADMIN_PASSWORD"`
	WsAllowedOrigins      string        `mapstructure:"WS_ALLOWED_ORIGINS"`
	WsTicketTTL           time.Duration `mapstructure:"WS_TICKET_TTL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.SetDefault("WS_TICKET_TTL", "1m")

	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Bind all environment variables explicitly
	// This works for both file-based config and environment variables
	envVars := []string{
		"DB_SOURCE", "SERVER_ADDRESS", "ACCESS_TOKEN_SECRET_KEY",
		"ACCESS_TOKEN_DURATION", "REFRESH_TOKEN_SECRET_KEY",
		"REFRESH_TOKEN_DURATION", "TWO_FA_TOKEN_SECRET_KEY",
		"TWO_FA_TOKEN_DURATION", "B2_ENDPOINT", "B2_KEY", "B2_KEY_ID", "B2_BUCKET",
		"DISABLE_BUCKET", "HOST", "REDIS_HOST", "REDIS_PASSWORD", "REMOTE",
		"OPEN_ROUTER_API_KEY", "SMTP_NAME", "SMTP_ADDRESS",
		"SMTP_AUTH", "SMTP_HOST", "SMTP_PORT", "BREVO_SENDER_NAME",
		"BREVO_SENDER_EMAIL", "BREVO_API_KEY", "ENVIRONMENT",
		"OPEN_ROUTER_API_KEY", "OPEN_ROUTER_MODEL",
		"MIGRATIONS_PATH", "ADMIN_EMAIL", "ADMIN_PASSWORD",
		"WS_ALLOWED_ORIGINS", "WS_TICKET_TTL",
	}

	for _, envVar := range envVars {
		err = viper.BindEnv(envVar)
		if err != nil {
			return config, fmt.Errorf("failed to bind event")
		}
	}

	// Try to read the config file (works locally)
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found (normal in Docker), using environment variables only
		} else {
			// Config file was found but another error occurred
			return config, err
		}
	}

	// Unmarshal the configuration
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Validate the configuration
	err = validateConfig(&config)
	if err != nil {
		return config, fmt.Errorf("invalid configuration: %w", err)
	}
	return config, err
}

func validateConfig(config *Config) error {
	// Define crucial environment variables that must be present
	crucialVars := map[string]string{
		"DB_SOURCE":                config.DbSource,
		"SERVER_ADDRESS":           config.ServerAddress,
		"ACCESS_TOKEN_SECRET_KEY":  config.AccessTokenSecretKey,
		"REFRESH_TOKEN_SECRET_KEY": config.RefreshTokenSecretKey,
		"TWO_FA_TOKEN_SECRET_KEY":  config.TwoFATokenSecretKey,
		"TWO_FA_TOKEN_DURATION":    config.TwoFATokenDuration.String(),
		"HOST":                     config.Host,
		"ENVIRONMENT":              config.Environment,
		"MIGRATIONS_PATH":          config.MigrationsPath,
	}

	var missingVars []string

	// Check if crucial variables are empty
	for varName, varValue := range crucialVars {
		if strings.TrimSpace(varValue) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	// Check duration fields (they should be greater than 0)
	if config.AccessTokenDuration <= 0 {
		missingVars = append(missingVars, "ACCESS_TOKEN_DURATION")
	}
	if config.RefreshTokenDuration <= 0 {
		missingVars = append(missingVars, "REFRESH_TOKEN_DURATION")
	}
	if config.TwoFATokenDuration <= 0 {
		missingVars = append(missingVars, "TWO_FA_TOKEN_DURATION")
	}
	if config.WsTicketTTL <= 0 {
		missingVars = append(missingVars, "WS_TICKET_TTL")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf(
			"missing or invalid crucial environment variables: %s",
			strings.Join(missingVars, ", "),
		)
	}

	return nil
}
