package plugins

import (
	"fmt"
	"time"

	"bot/client"
	"bot/messaging"
	"bot/messaging/helpers"
	btypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "runtime",
		Category: "System",
		FromMe:   false,
		IsGroup:  false,
		Handler:  Runtime,
	})
}

func Runtime(msg *events.Message, _ []string, sock *whatsmeow.Client) {
	uptime := helpers.FormatRuntime(time.Since(helpers.StartedAt))
	response := fmt.Sprintf("```\nRuntime: %s\n```", uptime)

	_, _ = client.SendMessage(client.SendOptions{
		JID:  msg.Info.Chat,
		Type: client.MsgText,
		Text: response,
	})
}
