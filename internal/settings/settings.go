package settings

import "database/sql"

// Get returns the value for a given key from the settings table.
func Get(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE `key`=?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// Set saves a key/value pair into the settings table.
func Set(db *sql.DB, key, value string) error {
	_, err := db.Exec("INSERT INTO settings(`key`,`value`) VALUES(?, ?) ON DUPLICATE KEY UPDATE `value`=VALUES(`value`)", key, value)
	return err
}
