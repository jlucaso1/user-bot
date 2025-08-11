package plugins

import (
	"bot/client"
	"bot/messaging"
	"bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&types.Command{
		Name:     "sticker",
		Category: "Media",
		FromMe:   true,
		IsGroup:  true,
		Handler:  Sticker,
	})
}

func Sticker(msg *events.Message, args []string, sock *whatsmeow.Client) {
	if len(args) == 0 {
		_, _ = client.SendMessage(types.SendOptions{
			JID:  msg.Info.Chat,
			Type: types.MsgText,
			Text: "please provide image file path",
		})
		return
	}

	stickerPath := args[0]
	Author := args[1]
	PackName := args[2]

	_, err := client.SendMessage(types.SendOptions{
		JID:      msg.Info.Chat,
		Type:     types.MsgSticker,
		FilePath: stickerPath,
		Author:   Author,
		PackName: PackName,
	})

	if err != nil {
		_, _ = client.SendMessage(types.SendOptions{
			JID:  msg.Info.Chat,
			Type: types.MsgText,
			Text: "failed to send sticker: " + err.Error(),
		})
	}
}
