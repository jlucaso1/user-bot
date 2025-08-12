package plugins

import (
	pk "bot/client"
	"bot/messaging"
	"bot/sql"
	"bot/types"
	"bot/utils"
	"strings"

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

	messaging.RegisterCommand(&types.Command{
		Name:     "setsudo",
		Category: "Settings",
		FromMe:   true,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			id := pk.GetUser(args, pk.ExtractContextInfo(msg.Message), msg.Info.Sender.String())
			user, err := pk.GetSender(id)
			if err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_no number found, mention or reply someone_"})
				return
			}

			isSudo, err := sql.IsSudo(user.PN)
			if err != nil {
				return
			}
			if isSudo {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_user is already sudo_"})
				return
			}

			if err := sql.SetSudo(user.PN, user.LID); err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_failed to save number in sudo, try again later_"})
				return
			}

			pk.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Type: types.MsgText,
				Text: "_sudo added_",
			})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "delsudo",
		Category: "Settings",
		FromMe:   true,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			id := pk.GetUser(args, pk.ExtractContextInfo(msg.Message), msg.Info.Sender.String())
			user, err := pk.GetSender(id)
			if err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_no number found_"})
				return
			}
			if err := sql.DelSudo(user.PN); err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_failed to delete sudo_"})
			}
			pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_sudo deleted_"})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "getsudo",
		Category: "Settings",
		FromMe:   true,
		Handler: func(msg *events.Message, args []string, client *whatsmeow.Client) {
			list, err := sql.ListSudo()

			if err != nil {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_failed to get sudo list_"})
				return
			}
			if len(list) == 0 {
				pk.SendMessage(types.SendOptions{JID: msg.Info.Chat, Type: types.MsgText, Text: "_no sudo numbers set_"})
				return
			}

			sudos := make([]string, len(list))
			copy(sudos, list)

			for i := range list {
				list[i] = "@" + strings.Split(list[i], "@")[0]
			}

			client.SendMessage(client.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String(strings.Join(list, "\n")),
					ContextInfo: &waE2E.ContextInfo{
						MentionedJID: sudos,
					},
				},
			})
		},
	})
}
