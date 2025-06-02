package util

import "fmt"

// ConstructDBConnectionString constructs a database connection string from the given parameters.
func ConstructDBConnectionString(user, pass, host, port, dbName string) string {
	// Construct a connection string with explicit parameters
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbName)
}
