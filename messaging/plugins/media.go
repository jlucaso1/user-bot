package plugins

import (
	"os"
	"path/filepath"
	"strings"

	"bot/client"
	"bot/messaging"
	"bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	messaging.RegisterCommand(&types.Command{
		Name:     "sticker",
		Category: "Media",
		FromMe:   false,
		Handler: func(msg *events.Message, args []string, sock *whatsmeow.Client) {
			if len(args) < 1 {
				sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("_please provide author|pack_"),
				})
				return
			}

			parts := strings.SplitN(args[0], "|", 2)
			if len(parts) < 2 {
				sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("_invalid format, use author|pack_"),
				})
				return
			}
			author, pack := parts[0], parts[1]

			ctx := client.ExtractContextInfo(msg.Message)
			if ctx.QuotedMessage == nil || (ctx.QuotedMessage.ImageMessage == nil && ctx.QuotedMessage.VideoMessage == nil) {
				sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
					Conversation: proto.String("_reply an image or video message_"),
				})
				return
			}

			var tmpFile string
			if ctx.QuotedMessage.ImageMessage != nil {
				data, err := sock.Download(sock.BackgroundEventCtx, ctx.QuotedMessage.ImageMessage)
				if err != nil {
					sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
						Conversation: proto.String("_failed to download media_"),
					})
					return
				}
				tmpFile = filepath.Join(os.TempDir(), "media.jpg")
				_ = os.WriteFile(tmpFile, data, 0644)
			} else if ctx.QuotedMessage.VideoMessage != nil {
				data, err := sock.Download(sock.BackgroundEventCtx, ctx.QuotedMessage.VideoMessage)
				if err != nil {
					sock.SendMessage(sock.BackgroundEventCtx, msg.Info.Chat, &waE2E.Message{
						Conversation: proto.String("_failed to download media_"),
					})
					return
				}
				tmpFile = filepath.Join(os.TempDir(), "media.mp4")
				_ = os.WriteFile(tmpFile, data, 0644)
			}

			client.SendMessage(types.SendOptions{
				JID:      msg.Info.Chat,
				Type:     types.MsgSticker,
				FilePath: tmpFile,
				Author:   author,
				PackName: pack,
			})
		},
	})
}
