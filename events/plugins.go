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

var commandRegex = regexp.MustCompile(`(?i)^[^\w\s]*([a-z0-9_]+)`)

func Plugins(sock *whatsmeow.Client, msg *events.Message) {
	if msg.Message == nil {
		return
	}

	messageText := utils.ExtractTextFromMessage(msg.Message)
	if messageText == "" {
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

	isSudo, err := sql.IsSudo(msg.Info.Sender.User)
	if err != nil {
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

	if cmd != nil {
		cmd.Handler(msg, args, sock)
		return
	}

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
}
