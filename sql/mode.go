package sql

import (
	"database/sql"
	"errors"
)

func syncMode() error {
	_, err := Conn.Exec(`
CREATE TABLE IF NOT EXISTS mode (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	value TEXT NOT NULL CHECK (value IN ('Public', 'Private'))
);
`)
	if err != nil {
		return err
	}

	_, err = Conn.Exec(`INSERT OR IGNORE INTO mode (id, value) VALUES (1, 'Private')`)
	return err
}

func GetMode() (string, error) {
	if Conn == nil {
		return "", errors.New("db connection not initialized")
	}

	if err := syncMode(); err != nil {
		return "", err
	}

	var mode string
	err := Conn.QueryRow(`SELECT value FROM mode WHERE id = 1`).Scan(&mode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return mode, nil
}

func SetMode(newMode string) error {
	if Conn == nil {
		return errors.New("db connection not initialized")
	}

	if newMode != "Public" && newMode != "Private" {
		return errors.New("invalid mode value")
	}

	_, err := Conn.Exec(`UPDATE mode SET value = ? WHERE id = 1`, newMode)
	return err
}
