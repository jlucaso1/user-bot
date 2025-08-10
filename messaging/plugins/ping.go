package plugins

import (
	"fmt"
	"time"

	"bot/client"
	"bot/messaging"
	btypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "ping",
		Category: "System",
		FromMe:   false,
		IsGroup:  false,
		Handler:  Ping,
	})
}

func Ping(msg *events.Message, _ []string, sock *whatsmeow.Client) {
	start := time.Now()

	res, _ := sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
		Conversation: proto.String("🏓 Pong!"),
	})

	duration := time.Since(start)
	response := fmt.Sprintf("```Pong (%v)```", duration.Round(time.Millisecond))

	_, _ = client.SendMessage(client.SendOptions{
		JID:        msg.Info.Chat,
		Type:       client.MsgEdit,
		MessageID:  res.ID,
		NewMessage: &waE2E.Message{Conversation: &response},
	})
}
