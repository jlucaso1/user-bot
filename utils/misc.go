package utils

import (
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
)

var fancyMap = map[rune]rune{
	'a': 'ᴀ', 'b': 'ʙ', 'c': 'ᴄ', 'd': 'ᴅ', 'e': 'ᴇ',
	'f': 'ғ', 'g': 'ɢ', 'h': 'ʜ', 'i': 'ɪ', 'j': 'ᴊ',
	'k': 'ᴋ', 'l': 'ʟ', 'm': 'ᴍ', 'n': 'ɴ', 'o': 'ᴏ',
	'p': 'ᴘ', 'q': 'ǫ', 'r': 'ʀ', 's': 's', 't': 'ᴛ',
	'u': 'ᴜ', 'v': 'ᴠ', 'w': 'ᴡ', 'x': 'x', 'y': 'ʏ', 'z': 'ᴢ',
}

func FancyText(input string) string {
	var out strings.Builder
	for _, r := range input {
		if r >= 'a' && r <= 'z' {
			if fr, ok := fancyMap[r]; ok {
				out.WriteRune(fr)
			} else {
				out.WriteRune(r)
			}
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func ExtractTextFromMessage(msg *waE2E.Message) string {
	switch {
	case msg.GetConversation() != "":
		return msg.GetConversation()
	case msg.ExtendedTextMessage != nil:
		return msg.ExtendedTextMessage.GetText()
	case msg.ImageMessage != nil && msg.ImageMessage.Caption != nil:
		return msg.ImageMessage.GetCaption()
	case msg.VideoMessage != nil && msg.VideoMessage.Caption != nil:
		return msg.VideoMessage.GetCaption()
	case msg.ProtocolMessage != nil && msg.ProtocolMessage.EditedMessage != nil:
		return ExtractTextFromMessage(msg.ProtocolMessage.EditedMessage)
	}
	return ""
}

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func Join(arr []string, sep string) string {
	return strings.Join(arr, sep)
}

func Map[T any, R any](arr []T, fn func(T) R) []R {
	out := make([]R, len(arr))
	for i, v := range arr {
		out[i] = fn(v)
	}
	return out
}

func ForEach[T any](arr []T, fn func(T)) {
	for _, v := range arr {
		fn(v)
	}
}

func Filter[T any](arr []T, fn func(T) bool) []T {
	out := []T{}
	for _, v := range arr {
		if fn(v) {
			out = append(out, v)
		}
	}
	return out
}

func Some[T any](arr []T, fn func(T) bool) bool {
	for _, v := range arr {
		if fn(v) {
			return true
		}
	}
	return false
}

func All[T any](arr []T, fn func(T) bool) bool {
	for _, v := range arr {
		if !fn(v) {
			return false
		}
	}
	return true
}

func Reduce[T any, R any](arr []T, fn func(R, T) R, initial R) R {
	acc := initial
	for _, v := range arr {
		acc = fn(acc, v)
	}
	return acc
}

func Find[T any](arr []T, fn func(T) bool) (T, bool) {
	for _, v := range arr {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func Includes[T comparable](arr []T, val T) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
