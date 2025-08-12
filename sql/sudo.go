package sql

import (
	"database/sql"
	"errors"
	"strings"
)

func syncSudo() error {
	_, err := Conn.Exec(`
CREATE TABLE IF NOT EXISTS sudo (
		jid TEXT PRIMARY KEY,
		lid TEXT NOT NULL
);
`)
	return err
}

func SetSudo(id1, id2 string) error {
	if Conn == nil {
		return errors.New("db connection not initialized")
	}
	if err := syncSudo(); err != nil {
		return err
	}

	var jid, lid string
	if strings.HasSuffix(id1, "@s.whatsapp.net") {
		jid = id1
		lid = id2
	} else if strings.HasSuffix(id2, "@s.whatsapp.net") {
		jid = id2
		lid = id1
	} else {
		return errors.New("invalid jid: must end with @s.whatsapp.net")
	}

	if !strings.HasPrefix(lid, "@") {
		lid = "@" + lid
	}

	_, err := Conn.Exec(`INSERT OR REPLACE INTO sudo (jid, lid) VALUES (?, ?)`, jid, lid)
	return err
}

func GetSudo(jid string) (string, string, error) {
	if Conn == nil {
		return "", "", errors.New("db connection not initialized")
	}
	if err := syncSudo(); err != nil {
		return "", "", err
	}

	var outJid, outLid string
	err := Conn.QueryRow(`SELECT jid, lid FROM sudo WHERE jid = ?`, jid).Scan(&outJid, &outLid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return outJid, outLid, nil
}

func DelSudo(jid string) error {
	if Conn == nil {
		return errors.New("db connection not initialized")
	}
	if err := syncSudo(); err != nil {
		return err
	}
	_, err := Conn.Exec(`DELETE FROM sudo WHERE jid = ?`, jid)
	return err
}

func IsSudo(value string) (bool, error) {
	if Conn == nil {
		return false, errors.New("db connection not initialized")
	}
	if err := syncSudo(); err != nil {
		return false, err
	}

	var count int
	err := Conn.QueryRow(`SELECT COUNT(*) FROM sudo WHERE jid = ? OR lid = ?`, value, value).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
