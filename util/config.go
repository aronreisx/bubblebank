package util

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBImage          string `mapstructure:"DB_IMAGE"`
	DBVersion        string `mapstructure:"DB_VERSION"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBUser           string `mapstructure:"DB_USER"`
	DBPass           string `mapstructure:"DB_PASS"`
	DBName           string `mapstructure:"DB_NAME"`
	DBHost           string `mapstructure:"DB_HOST"`
	ServerPort       string `mapstructure:"SERVER_PORT"`
	MigrationsFolder string `mapstructure:"MIGRATIONS_FOLDER"`

	// OpenTelemetry Configuration
	OtelServiceName          string `mapstructure:"OTEL_SERVICE_NAME"`
	OtelServiceVersion       string `mapstructure:"OTEL_SERVICE_VERSION"`
	OtelEnvironment          string `mapstructure:"OTEL_ENVIRONMENT"`
	OtelExporterOtlpEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`

	// Observability Configuration
	PrometheusEndpoint string `mapstructure:"PROMETHEUS_ENDPOINT"`
	MetricsPort        string `mapstructure:"METRICS_PORT"`
	LogLevel           string `mapstructure:"LOG_LEVEL"`
}

// LoadConfig reads configuration from file or OS variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Important: Set environment variable name mapping
	viper.SetEnvPrefix("") // No prefix needed since env vars already have proper names

	// List of environment variables to bind
	envVars := []string{
		"DB_IMAGE",
		"DB_VERSION",
		"DB_PORT",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"DB_HOST",
		"SERVER_PORT",
		"MIGRATIONS_FOLDER",
		// OpenTelemetry
		"OTEL_SERVICE_NAME",
		"OTEL_SERVICE_VERSION",
		"OTEL_ENVIRONMENT",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		// Observability
		"PROMETHEUS_ENDPOINT",
		"METRICS_PORT",
		"LOG_LEVEL",
	}

	// Map all environment variables to viper keys in a loop
	for _, envVar := range envVars {
		if err := viper.BindEnv(envVar); err != nil {
			return Config{}, fmt.Errorf("failed to bind environment variable %s: %w", envVar, err)
		}
	}

	// Enable automatic environment variable lookup
	viper.AutomaticEnv()

	// New implementation
	// Attempt to read the config file
	if err = viper.ReadInConfig(); err != nil {
		log.Printf("Config warning: %v. Using environment variables instead.", err)
	}

	// Check for missing required variables
	var missingVars []string
	for _, envName := range envVars {
		if viper.GetString(envName) == "" {
			missingVars = append(missingVars, envName)
		}
	}

	// Log missing environment variables
	if len(missingVars) > 0 {
		log.Printf(
			"Config warning: Environment variables %v are missing or empty.",
			strings.Join(missingVars, ", "),
		)
	}

	// Unmarshal the config file or environment variables
	if err = viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
