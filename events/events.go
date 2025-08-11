package events

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func EventHandler(sock *whatsmeow.Client, evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		Plugins(sock, v)
	case *events.Connected:
		// sock.SendMessage(sock.BackgroundEventCtx, *sock.Store.ID, &waE2E.Message{
		// 	Conversation: proto.String("```whatsmeow connected```"),
		// })
	}
}
