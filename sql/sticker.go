package sql

import (
	"database/sql"
	"errors"
	"fmt"
)

func syncStickerCmd() error {
	_, err := Conn.Exec(`
CREATE TABLE IF NOT EXISTS stickercmd (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	cmd TEXT UNIQUE NOT NULL,
	value TEXT NOT NULL
);
`)
	return err
}

func SetStickerCmd(cmd string, stickerBytes []byte) error {
	if Conn == nil {
		return errors.New("db connection not initialized")
	}

	if err := syncStickerCmd(); err != nil {
		return err
	}

	value := fmt.Sprintf("%x", stickerBytes)

	_, err := Conn.Exec(`
INSERT INTO stickercmd (cmd, value)
VALUES (?, ?)
ON CONFLICT(cmd) DO UPDATE SET value = excluded.value
`, cmd, value)

	return err
}

func GetStickerCmd(cmd string) (string, error) {
	if Conn == nil {
		return "", errors.New("db connection not initialized")
	}

	if err := syncStickerCmd(); err != nil {
		return "", err
	}

	var value string
	err := Conn.QueryRow(`SELECT value FROM stickercmd WHERE cmd = ?`, cmd).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return value, nil
}

func DelStickerCmd(cmd string) error {
	if Conn == nil {
		return errors.New("db connection not initialized")
	}

	if err := syncStickerCmd(); err != nil {
		return err
	}

	_, err := Conn.Exec(`DELETE FROM stickercmd WHERE cmd = ?`, cmd)
	return err
}

func ListStickerCmds() ([]string, error) {
	if Conn == nil {
		return nil, errors.New("db connection not initialized")
	}

	if err := syncStickerCmd(); err != nil {
		return nil, err
	}

	rows, err := Conn.Query(`SELECT cmd FROM stickercmd`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			continue
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}
