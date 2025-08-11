package plugins

import (
	"bot/client"
	"bot/messaging"
	btypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "repo",
		FromMe:   false,
		Category: "Utils",
		Handler: func(msg *events.Message, _ []string, sock *whatsmeow.Client) {
			_, _ = client.SendMessage(btypes.SendOptions{
				JID:      msg.Info.Chat,
				Type:     btypes.MsgImage,
				FilePath: "./resources/logo.png",
				Caption:  "Simple User WhatsAppBot\nhttps://github.com/AstroX11/user-bot",
			})
		},
	})
}
