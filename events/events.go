package events

import (
	"bot/client"
	"bot/sql"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

var connectedSent bool
var mu sync.Mutex

func EventHandler(sock *whatsmeow.Client, evt interface{}) {
	mu.Lock()
	if !connectedSent {
		if sock.Store.ID != nil {
			jid := sock.Store.ID.ToNonAD()
			lid := sock.Store.LID.ToNonAD().String()
			
			parts := strings.SplitN(jid.String(), "@", 2)

			msg := "bot connected\n" +
				"go version: " + runtime.Version() + "\n" +
				"goroutines: " + strconv.Itoa(runtime.NumGoroutine()) + "\n" +
				"user: @" + parts[0]

			_, err := sock.SendMessage(sock.BackgroundEventCtx, jid, &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String("```" + msg + "```"),
					ContextInfo: &waE2E.ContextInfo{
						MentionedJID: []string{jid.String()},
					},
				},
			})
			sql.SetSudo(jid.String(), lid)
			if err == nil {
				connectedSent = true
			}
		}
	}
	mu.Unlock()

	switch evt := evt.(type) {
	case *events.Message:
		client.SaveSender(client.CleanID(evt.Info.Sender.String()), client.CleanID(evt.Info.SenderAlt.String()))
		Plugins(sock, evt)
	}
}
