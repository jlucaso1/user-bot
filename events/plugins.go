package events

import (
	"fmt"
	"regexp"
	"strings"

	"bot/client"
	"bot/messaging"
	"bot/sql"
	"bot/types"
	"bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

var commandRegex = regexp.MustCompile(`(?i)^[^\w\s]*\s*([a-z0-9_]+)`)

func safeExecute(handler func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[panic recovered] %v\n", r)
		}
	}()
	handler()
}

func Plugins(sock *whatsmeow.Client, msg *events.Message) {
	if sock == nil || msg == nil || msg.Message == nil {
		return
	}

	messageText := utils.ExtractTextFromMessage(msg.Message)
	if strings.TrimSpace(messageText) == "" {
		return
	}

	prefix, err := sql.GetPrefix()
	if err != nil || prefix == "" {
		prefix = "."
	}

	messageText = strings.TrimSpace(messageText)
	if !strings.HasPrefix(messageText, prefix) {
		return
	}

	match := commandRegex.FindStringSubmatch(messageText)
	if len(match) < 2 {
		return
	}
	cmdName := strings.ToLower(match[1])
	args := strings.Fields(messageText)[1:]

	cmd := messaging.FindCommand(cmdName)
	if cmd == nil {
		suggestion := messaging.SuggestCommand(cmdName)
		if suggestion != "" {
			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Text: fmt.Sprintf("❌ Command `%s` not found. Did you mean `%s%s`?", cmdName, prefix, suggestion),
			})
		} else {
			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Text: fmt.Sprintf("❌ Command `%s` not found.", cmdName),
			})
		}
		return
	}

	isSudo, err := sql.IsSudo(msg.Info.Sender.ToNonAD().String())
	if err != nil {
		return
	}

	mode, err := sql.GetMode()
	if err != nil {
		return
	}

	if mode == "Private" && !isSudo {
		return
	}

	if cmd.FromMe && !isSudo {
		sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			Conversation: proto.String("_this command is for sudo users_"),
		})
		return
	}

	if cmd.IsGroup && !msg.Info.IsGroup {
		sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			Conversation: proto.String("_this command is for groups_"),
		})
		return
	}

	safeExecute(func() {
		cmd.Handler(msg, args, sock)
	})
}

func Sticker(sock *whatsmeow.Client, msg *events.Message) {
	if sock == nil || msg == nil || msg.Message == nil || msg.Message.StickerMessage == nil {
		return
	}

	stickerHash := fmt.Sprintf("%x", msg.Message.StickerMessage.FileSHA256)
	if stickerHash == "" {
		return
	}

	rows, err := sql.Conn.Query(`SELECT cmd FROM stickercmd WHERE value = ?`, stickerHash)
	if err != nil {
		return
	}
	defer rows.Close()

	var cmdName string
	if rows.Next() {
		if err := rows.Scan(&cmdName); err != nil {
			return
		}
	} else {
		return
	}

	cmd := messaging.FindCommand(cmdName)
	if cmd == nil {
		return
	}

	isSudo, err := sql.IsSudo(msg.Info.Sender.ToNonAD().String())
	if err != nil {
		return
	}

	mode, err := sql.GetMode()
	if err != nil {
		return
	}

	if mode == "Private" && !isSudo {
		return
	}

	if cmd.FromMe && !isSudo {
		sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			Conversation: proto.String("_this command is for sudo users_"),
		})
		return
	}

	if cmd.IsGroup && !msg.Info.IsGroup {
		sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			Conversation: proto.String("_this command is for groups_"),
		})
		return
	}

	safeExecute(func() {
		cmd.Handler(msg, []string{}, sock)
	})
}
