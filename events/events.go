package events

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	"runtime"
	"strconv"
	"sync"
)

var connectedSent bool
var mu sync.Mutex

func EventHandler(sock *whatsmeow.Client, evt interface{}) {
	mu.Lock()
	if !connectedSent {
		msg := "bot connected\n" +
			"go version: " + runtime.Version() + "\n" +
			"goroutines: " + strconv.Itoa(runtime.NumGoroutine()) + "\n" +
			"user: " + sock.Store.ID.String()
		_, _ = sock.SendMessage(sock.BackgroundEventCtx, *sock.Store.ID, &waE2E.Message{
			Conversation: proto.String("```" + msg + "```"),
		})
		connectedSent = true
	}
	mu.Unlock()

	switch evt := evt.(type) {
	case *events.Message:
		Plugins(sock, evt)
	}
}
