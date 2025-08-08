package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a MySQL connection using the MYSQL_DSN environment variable.
// Example DSN: user:password@tcp(localhost:3306)/newsdb?parseTime=true&charset=utf8mb4
func Connect() (*sql.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("MYSQL_DSN not set")
	}
	return sql.Open("mysql", dsn)
}
