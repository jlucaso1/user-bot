package events

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func EventHandler(sock *whatsmeow.Client, evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		Plugins(sock, v)
	}
}
