package util

import (
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
	}

	// Map all environment variables to viper keys in a loop
	for _, envVar := range envVars {
		viper.BindEnv(envVar)
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
