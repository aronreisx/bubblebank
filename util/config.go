package util

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBImage    string `mapstructure:"DB_IMAGE"`
	DBVersion  string `mapstructure:"DB_VERSION"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPass     string `mapstructure:"DB_PASS"`
	DBName     string `mapstructure:"DB_NAME"`
	ServerPort string `mapstructure:"SERVER_PORT"`
}

// LoadConfig reads configuration from file or OS variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Attempt to read the config file
	if err = viper.ReadInConfig(); err != nil {
		log.Printf("Config warning: %v. Falling back to environment variables.", err)

		// List of environment variables to set defaults for
		envVars := []string{
			"DB_IMAGE",
			"DB_VERSION",
			"DB_PORT",
			"DB_USER",
			"DB_PASS",
			"DB_NAME",
			"SERVER_PORT",
		}

		// Set default values from environment variables
		var missingVars []string
		for _, envName := range envVars {
			osEnvVarValue := os.Getenv(envName)

			// Check if any environment variables are missing
			if osEnvVarValue == "" {
				missingVars = append(missingVars, envName)
			}

			viper.SetDefault(envName, osEnvVarValue)
		}

		// Log missing environment variables
		if len(missingVars) > 0 {
			log.Printf(
				"Config warning: Environment variables %v are missing.",
				strings.Join(missingVars, ", "),
			)
		}
	}

	// Unmarshal the config file or environment variables
	if err = viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
