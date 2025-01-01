package util

import "fmt"

// ConstructDBUrl constructs a database URL from the given parameters.
func ConstructDBUrl(user, pass, host, port, dbName string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, dbName,
	)
}
