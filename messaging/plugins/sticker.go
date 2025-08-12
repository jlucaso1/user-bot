package plugins

import (
	"bot/client"
	"bot/messaging"
	"bot/sql"
	"bot/types"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&types.Command{
		Name:     "setcmd",
		Category: "Misc",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, sock *whatsmeow.Client) {
			quoted := client.ExtractContextInfo(msg.Message)
			if quoted == nil || quoted.QuotedMessage == nil || quoted.QuotedMessage.StickerMessage == nil {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_reply a sticker message_",
				})
				return
			}
			if len(args) == 0 {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_please provide a command name_",
				})
				return
			}
			cmd := args[0]
			if !messaging.IsCommand(cmd) {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_" + cmd + " is not a valid command_",
				})
				return
			}

			err := sql.SetStickerCmd(cmd, quoted.QuotedMessage.StickerMessage.FileSHA256)
			if err != nil {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_failed to set sticker cmd_",
				})
				return
			}

			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Type: types.MsgText,
				Text: "_sticker cmd set for " + cmd + " command_",
			})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "delcmd",
		Category: "Misc",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, sock *whatsmeow.Client) {
			if len(args) == 0 {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_please provide a command to delete_",
				})
				return
			}
			cmd := args[0]
			err := sql.DelStickerCmd(cmd)
			if err != nil {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_failed to delete sticker cmd_",
				})
				return
			}
			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Type: types.MsgText,
				Text: "_deleted sticker cmd for " + cmd + "_",
			})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "getcmd",
		Category: "Misc",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, sock *whatsmeow.Client) {
			if len(args) == 0 {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_please provide a command to get_",
				})
				return
			}
			cmd := args[0]
			value, err := sql.GetStickerCmd(cmd)
			if err != nil {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_failed to get sticker cmd_",
				})
				return
			}
			if value == "" {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_no sticker command found for " + cmd + "_",
				})
				return
			}
			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Type: types.MsgText,
				Text: fmt.Sprintf("_sticker command for %s: %s_", cmd, value),
			})
		},
	})

	messaging.RegisterCommand(&types.Command{
		Name:     "listcmd",
		Category: "Misc",
		FromMe:   true,
		IsGroup:  false,
		Handler: func(msg *events.Message, args []string, sock *whatsmeow.Client) {
			cmds, err := sql.ListStickerCmds()
			if err != nil {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_failed to list sticker commands_",
				})
				return
			}
			if len(cmds) == 0 {
				client.SendMessage(types.SendOptions{
					JID:  msg.Info.Chat,
					Type: types.MsgText,
					Text: "_no sticker commands set_",
				})
				return
			}

			text := "sticker commands:\n"
			for _, c := range cmds {
				text += "- " + c + "\n"
			}

			client.SendMessage(types.SendOptions{
				JID:  msg.Info.Chat,
				Type: types.MsgText,
				Text: text,
			})
		},
	})
}
