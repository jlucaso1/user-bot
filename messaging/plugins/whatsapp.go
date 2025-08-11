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

	// messaging.RegisterCommand(&types.Command{
	// 	Name:     "pp",
	// 	Category: "WhatsApp",
	// 	FromMe:   true,
	// 	IsGroup:  false,
	// 	Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
	// 		contextInfo := pk.ExtractContextInfo(msg.Message)

	// 		if contextInfo == nil {
	// 			_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
	// 				Conversation: proto.String("reply an image"),
	// 			})
	// 			return
	// 		}

	// 		quotedMsg := contextInfo.QuotedMessage
	// 		if quotedMsg == nil || quotedMsg.ImageMessage == nil {
	// 			_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
	// 				Conversation: proto.String("reply an image"),
	// 			})
	// 			return
	// 		}

	// 		avatar, err := client.Download(client.BackgroundEventCtx, quotedMsg.ImageMessage)
	// 		if err != nil {
	// 			_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
	// 				Conversation: proto.String("failed to download image"),
	// 			})
	// 			return
	// 		}

	// 		client.SetGroupPhoto(*client.Store.ID, avatar)
	// 		pk.SendMessage(types.SendOptions{
	// 			JID:  *client.Store.ID,
	// 			Text: *proto.String("profile photo updated"),
	// 		})
	// 	},
	// })

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
			// var media []byte
			// var err error
			// var msgType types.MessageType
			// var ext string

			// switch {
			// case qMsg.ImageMessage != nil && qMsg.ImageMessage.ViewOnce != nil && *qMsg.ImageMessage.ViewOnce:
			// 	media, err = client.Download(client.BackgroundEventCtx, qMsg.ImageMessage)
			// 	msgType = types.MsgImage
			// 	ext = ".jpg"
			// case qMsg.VideoMessage != nil && qMsg.VideoMessage.ViewOnce != nil && *qMsg.VideoMessage.ViewOnce:
			// 	media, err = client.Download(client.BackgroundEventCtx, qMsg.VideoMessage)
			// 	msgType = types.MsgVideo
			// 	ext = ".mp4"
			// case qMsg.AudioMessage != nil && qMsg.AudioMessage.ViewOnce != nil && *qMsg.AudioMessage.ViewOnce:
			// 	media, err = client.Download(client.BackgroundEventCtx, qMsg.AudioMessage)
			// 	msgType = types.MsgAudio
			// 	ext = ".opus"
			// default:
			// 	_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			// 		Conversation: proto.String("reply a valid viewonce image, video, or audio"),
			// 	})
			// 	return
			// }

			// if err != nil {
			// 	_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			// 		Conversation: proto.String("failed to download media"),
			// 	})
			// 	return
			// }

			// tmpFile, err := os.CreateTemp("", "viewonce-*"+ext)
			// if err != nil {
			// 	_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			// 		Conversation: proto.String("failed to create temp file"),
			// 	})
			// 	return
			// }
			// defer tmpFile.Close()

			// _, err = tmpFile.Write(media)
			// if err != nil {
			// 	_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			// 		Conversation: proto.String("failed to write temp file"),
			// 	})
			// 	return
			// }

			// caption := ""
			// if qMsg.ImageMessage != nil {
			// 	caption = qMsg.ImageMessage.GetCaption()
			// } else if qMsg.VideoMessage != nil {
			// 	caption = qMsg.VideoMessage.GetCaption()
			// } else {
			// 	caption = ""
			// }

			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, qMsg)

			// pk.SendMessage(types.SendOptions{
			// 	JID:      msg.Info.Chat,
			// 	Type:     msgType,
			// 	FilePath: tmpFile.Name(),
			// 	Caption:  caption,
			// 	IsVoiceNote: true,
			// })
			// if err != nil {
			// 	_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
			// 		Conversation: proto.String("failed to send media"),
			// 	})
			// 	return
			// }
		},
	})

}
