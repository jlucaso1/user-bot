package events

import (
	"bot/utils"

	"go.mau.fi/whatsmeow/types/events"
)

func EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		utils.LogPretty(v)
		Plugins(v)

	}
}
