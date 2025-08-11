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
		Name:     "sticker",
		Category: "Media",
		FromMe:   true,
		IsGroup:  true,
		Handler:  Sticker,
	})
}

func Sticker(msg *events.Message, args []string, sock *whatsmeow.Client) {
	if len(args) == 0 {
		_, _ = client.SendMessage(client.SendOptions{
			JID:  msg.Info.Chat,
			Type: client.MsgText,
			Text: "please provide image file path",
		})
		return
	}

	stickerPath := args[0]
	Author := args[1]
	PackName := args[2]

	_, err := client.SendMessage(client.SendOptions{
		JID:      msg.Info.Chat,
		Type:     client.MsgSticker,
		FilePath: stickerPath,
		Author:   Author,
		PackName: PackName,
	})

	if err != nil {
		_, _ = client.SendMessage(client.SendOptions{
			JID:  msg.Info.Chat,
			Type: client.MsgText,
			Text: "failed to send sticker: " + err.Error(),
		})
	}
}
