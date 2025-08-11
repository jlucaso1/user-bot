package types

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type MessageType string

const (
	MsgText    MessageType = "text"
	MsgImage   MessageType = "image"
	MsgVideo   MessageType = "video"
	MsgDoc     MessageType = "document"
	MsgAudio   MessageType = "audio"
	MsgSticker MessageType = "sticker"
	MsgEdit    MessageType = "edit"
)

type SendOptions struct {
	JID         types.JID
	Type        MessageType
	Text        string
	FilePath    string
	FileName    string
	Caption     string
	IsVoiceNote bool
	MessageID   string
	NewMessage  *waE2E.Message
	Author      string
	PackName    string
	Categories  []string
}
