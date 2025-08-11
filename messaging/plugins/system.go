package plugins

import (
	"fmt"
	"time"

	"bot/client"
	"bot/messaging"
	"bot/messaging/helpers"
	btypes "bot/types"
	"bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "runtime",
		Category: "System",
		FromMe:   false,
		IsGroup:  false,
		Handler: func(msg *events.Message, _ []string, sock *whatsmeow.Client) {
			uptime := helpers.FormatRuntime(time.Since(helpers.StartedAt))
			response := fmt.Sprintf("```\nRuntime: %s\n```", uptime)

			_, _ = client.SendMessage(btypes.SendOptions{
				JID:  msg.Info.Chat,
				Type: btypes.MsgText,
				Text: response,
			})
		},
	})

	messaging.RegisterCommand(&btypes.Command{

		Name:     "restart",
		Category: "System",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			utils.Restart()
		},
	})

	messaging.RegisterCommand(&btypes.Command{
		Name:     "ping",
		Category: "System",
		FromMe:   false,
		IsGroup:  false,
		Handler: func(msg *events.Message, _ []string, sock *whatsmeow.Client) {
			start := time.Now()

			res, _ := sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
				Conversation: proto.String("🏓 Pong!"),
			})

			duration := time.Since(start)
			response := fmt.Sprintf("```Pong (%v)```", duration.Round(time.Millisecond))

			_, _ = client.SendMessage(btypes.SendOptions{
				JID:        msg.Info.Chat,
				Type:       btypes.MsgEdit,
				MessageID:  res.ID,
				NewMessage: &waE2E.Message{Conversation: &response},
			})
		},
	})
}
