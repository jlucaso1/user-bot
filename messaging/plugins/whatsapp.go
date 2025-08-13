package plugins

import (
	pk "bot/client"
	"bot/messaging"
	"bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	watypes "go.mau.fi/whatsmeow/types"
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

	messaging.RegisterCommand(&types.Command{
		Name:     "tovv",
		Category: "WhatsApp",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			contextInfo := pk.ExtractContextInfo(msg.Message)
			if contextInfo == nil || contextInfo.QuotedMessage == nil {
				_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("reply an image, video, or audio to make it view once"),
				})
				return
			}

			qMsg := contextInfo.QuotedMessage

			switch {
			case qMsg.ImageMessage != nil:
				qMsg.ImageMessage.ViewOnce = proto.Bool(true)
			case qMsg.VideoMessage != nil:
				qMsg.VideoMessage.ViewOnce = proto.Bool(true)
			case qMsg.AudioMessage != nil:
				qMsg.AudioMessage.ViewOnce = proto.Bool(true)
			default:
				_, _ = client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("only image, video, or audio are supported"),
				})
				return
			}
			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, qMsg)
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "block",
		Category: "WhatsApp",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			var target string

			if !msg.Info.IsGroup {
				target = msg.Info.Chat.String()
			} else {
				id := pk.GetUser(args, pk.ExtractContextInfo(msg.Message), msg.Info.Sender.String())
				user, err := pk.GetSender(id)
				if err != nil || user.PN == client.Store.ID.ToNonAD().User {
					pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_no number found, mention or reply someone_"})
					return
				}
				target = user.PN
			}

			toBlock, err := watypes.ParseJID(target)
			if err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_invalid number_"})
				return
			}

			client.UpdateBlocklist(toBlock, "block")
			pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_blocked_"})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "unblock",
		Category: "WhatsApp",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			var target string

			if !msg.Info.IsGroup {
				target = msg.Info.Chat.String()
			} else {
				id := pk.GetUser(args, pk.ExtractContextInfo(msg.Message), msg.Info.Sender.String())
				user, err := pk.GetSender(id)
				if err != nil || user.PN == client.Store.ID.ToNonAD().User {
					pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_no number found, mention or reply someone_"})
					return
				}
				target = user.PN
			}

			toUnblock, err := watypes.ParseJID(target)
			if err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_invalid number_"})
				return
			}

			client.UpdateBlocklist(toUnblock, "unblock")
			pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_unblocked_"})
		},
	})

}
