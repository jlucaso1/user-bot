package plugins

import (
	"strings"

	"bot/client"
	"bot/messaging"
	btypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "audio",
		Category: "Test",
		FromMe:   true,
		IsGroup:  false,
		Handler:  SendTestAudio,
	})
}

func SendTestAudio(msg *events.Message, args []string, sock *whatsmeow.Client) {
	audioPath := "./resources/audio.mp3"

	isVoice := false
	if len(args) > 0 && (strings.EqualFold(args[0], "vn") || strings.EqualFold(args[0], "ptt")) {
		isVoice = true
	}

	_, _ = client.SendMessage(btypes.SendOptions{
		JID:         msg.Info.Chat,
		Type:        btypes.MsgAudio,
		FilePath:    audioPath,
		IsVoiceNote: isVoice,
	})
}
