package plugins

import (
	pk "bot/client"
	"bot/messaging"
	"bot/sql"
	"bot/types"
	"bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	messaging.RegisterCommand(&types.Command{
		Name:     "prefix",
		Category: "Settings",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			if len(args) < 1 || args[0] == "" {
				prefix, err := sql.GetPrefix()
				if err != nil {
					return
				}
				pk.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "usage: " + prefix + "prefix <symbol>",
				})
				return
			}

			sql.SetPrefix(args[0])

			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
				Conversation: proto.String("prefix set to " + args[0]),
			})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "mode",
		Category: "Settings",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			if len(args) < 1 {
				currentMode, err := sql.GetMode()
				if err != nil {
					return
				}
				pk.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "usage: mode <Public|Private>\ncurrent mode: " + currentMode,
				})
				return
			}

			mode := utils.Ucfirst(args[0])
			if mode != "Public" && mode != "Private" {
				pk.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "usage: mode <Public|Private>\ninvalid mode: " + args[0],
				})
				return
			}

			currentMode, err := sql.GetMode()
			if err != nil {
				return
			}

			if currentMode == mode {
				pk.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "mode is already set to " + mode,
				})
				return
			}

			err = sql.SetMode(mode)
			if err != nil {
				pk.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "failed to set mode: " + err.Error(),
				})
				return
			}

			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
				Conversation: proto.String("mode set to " + mode),
			})
		},
	})

}
