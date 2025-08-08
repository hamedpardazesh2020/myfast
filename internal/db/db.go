package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

// Open creates a database connection using the MYSQL_DSN environment variable.
// If the variable is empty, it falls back to a default DSN.
func Open() (*sql.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/newsdb?charset=utf8mb4&parseTime=True"
	}
	return sql.Open("mysql", dsn)
}
