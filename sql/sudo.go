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

	if !strings.HasSuffix(jid, "@s.whatsapp.net") {
		jid = clid(jid) + "@s.whatsapp.net"
	}
	if !strings.HasSuffix(lid, "@lid") {
		lid = clid(lid) + "@lid"
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

func ListSudo() ([]string, error) {
	if Conn == nil {
		return nil, errors.New("db connection not initialized")
	}
	if err := syncSudo(); err != nil {
		return nil, err
	}

	rows, err := Conn.Query(`SELECT jid FROM sudo`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sudos []string
	for rows.Next() {
		var jid string
		if err := rows.Scan(&jid); err != nil {
			return nil, err
		}
		sudos = append(sudos, jid)
	}

	return sudos, rows.Err()
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


func clid(s string) string {
	start := 0
	for start < len(s) {
		r := rune(s[start])
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			break
		}
		start++
	}
	s = s[start:]

	for i, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return s[:i]
		}
	}
	return s
}