package plugins

import (
	pk "bot/client"
	"bot/messaging"
	"bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	messaging.RegisterCommand(&types.Command{
		Name:     "logout",
		Category: "WhatsApp",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			client.Logout(client.BackgroundEventCtx)
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "vv",
		Category: "WhatsApp",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			contextInfo := pk.ExtractContextInfo(msg.Message)
			if contextInfo == nil || contextInfo.QuotedMessage == nil {
				_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("reply a viewonce image, video, or audio"),
				})
				return
			}

			qMsg := contextInfo.QuotedMessage

			switch {
			case qMsg.ImageMessage != nil:
				qMsg.ImageMessage.ViewOnce = proto.Bool(false)
			case qMsg.VideoMessage != nil:
				qMsg.VideoMessage.ViewOnce = proto.Bool(false)
			case qMsg.AudioMessage != nil:
				qMsg.AudioMessage.ViewOnce = proto.Bool(false)

			}
			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, qMsg)
		},
	})
}
