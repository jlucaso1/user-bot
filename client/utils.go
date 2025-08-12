package client

import (
	"bot/sql"
	"errors"
	"strings"

	botypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func IsAdmin(selfJID types.JID, participants []types.GroupParticipant) bool {
	for _, p := range participants {
		if p.JID.User == selfJID.User {
			return p.IsAdmin || p.IsSuperAdmin
		}
	}
	return false
}

func RevokeMessage(sock *whatsmeow.Client, chatID types.JID, messageID string, IsGroup bool) {}

func GetSender(id string) (*botypes.PNLID, error) {
	id = CleanID(id)

	var res botypes.PNLID

	err := sql.Conn.QueryRow(`
		SELECT lid, pn
		FROM whatsmeow_lid_map
		WHERE lid = ? OR pn = ?
		LIMIT 1
	`, id, id).Scan(&res.LID, &res.PN)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func SaveSender(a, b string) error {
	a = CleanID(a)
	b = CleanID(b)

	var pn, lid string

	if strings.HasSuffix(a, "@s.whatsapp.net") {
		pn = a
	} else if strings.HasSuffix(a, "@lid") {
		lid = a
	}

	if strings.HasSuffix(b, "@s.whatsapp.net") {
		pn = b
	} else if strings.HasSuffix(b, "@lid") {
		lid = b
	}

	if pn == "" && lid == "" {
		return errors.New("invalid sender values")
	}

	_, err := sql.Conn.Exec(`
		INSERT OR REPLACE INTO whatsmeow_lid_map (lid, pn)
		VALUES (?, ?)
	`, lid, pn)

	return err
}

func CleanID(s string) string {
	for i, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return s[:i]
		}
	}
	return s
}

func GetUser(args []string, ctx *waE2E.ContextInfo, Sender string) string {
	if len(args) > 0 && args[0] != "" && strings.Contains(args[0], "@") {
		return strings.ReplaceAll(args[0], "@", "")
	}
	if ctx != nil && ctx.GetParticipant() != "" {
		return ctx.GetParticipant()
	}
	if ctx != nil && len(ctx.MentionedJID) > 0 && ctx.MentionedJID[0] != "" {
		return ctx.MentionedJID[0]
	}
	return Sender
}
