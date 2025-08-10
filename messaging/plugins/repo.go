package plugins

import (
	"bot/client"
	"bot/messaging"
	btypes "bot/types"

	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "repo",
		FromMe:   false,
		Category: "misc",
		Handler: func(msg *events.Message, _ []string) {
			_, _ = client.SendMessage(client.SendOptions{
				JID:      msg.Info.Chat,
				Type:     client.MsgImage,
				FilePath: "./resources/logo.png",
				Caption:  "Simple User WhatsAppBot\nhttps://github.com/AstroX11/user-bot",
			})
		},
	})
}
