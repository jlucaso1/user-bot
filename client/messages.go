package client

import (
	"bot/types"
	"bot/utils"
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)


var sock *whatsmeow.Client

func SetClient(c *whatsmeow.Client) { sock = c }

func SendMessage(opts types.SendOptions) (string, error) {
	if sock == nil {
		return "", fmt.Errorf("client not initialized")
	}

	switch opts.Type {
	case types.MsgText:
		msg := &waE2E.Message{Conversation: proto.String(opts.Text)}
		resp, err := sock.SendMessage(context.Background(), opts.JID, msg)
		if err != nil {
			return "", err
		}
		return resp.ID, nil

	case types.MsgImage:
		return sendMedia(opts, whatsmeow.MediaImage, func(urls uploadedMedia, size int64) *waE2E.Message {
			mimeType := mime.TypeByExtension(filepath.Ext(opts.FilePath))
			if mimeType == "" {
				mimeType = "image/jpeg"
			}
			return &waE2E.Message{
				ImageMessage: &waE2E.ImageMessage{
					URL:           proto.String(urls.URL),
					DirectPath:    proto.String(urls.DirectPath),
					MediaKey:      urls.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSHA256: urls.FileEncSHA256,
					FileSHA256:    urls.FileSHA256,
					FileLength:    proto.Uint64(uint64(size)),
					Caption:       proto.String(opts.Caption),
				},
			}
		})

	case types.MsgVideo:
		return sendMedia(opts, whatsmeow.MediaVideo, func(urls uploadedMedia, size int64) *waE2E.Message {
			mimeType := mime.TypeByExtension(filepath.Ext(opts.FilePath))
			if mimeType == "" {
				mimeType = "video/mp4"
			}
			return &waE2E.Message{
				VideoMessage: &waE2E.VideoMessage{
					URL:           proto.String(urls.URL),
					DirectPath:    proto.String(urls.DirectPath),
					MediaKey:      urls.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSHA256: urls.FileEncSHA256,
					FileSHA256:    urls.FileSHA256,
					FileLength:    proto.Uint64(uint64(size)),
					Caption:       proto.String(opts.Caption),
				},
			}
		})

	case types.MsgDoc:
		return sendMedia(opts, whatsmeow.MediaDocument, func(urls uploadedMedia, size int64) *waE2E.Message {
			fileName := opts.FileName
			if fileName == "" {
				fileName = filepath.Base(opts.FilePath)
			}
			mimeType := mime.TypeByExtension(filepath.Ext(opts.FilePath))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			return &waE2E.Message{
				DocumentMessage: &waE2E.DocumentMessage{
					URL:           proto.String(urls.URL),
					DirectPath:    proto.String(urls.DirectPath),
					MediaKey:      urls.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSHA256: urls.FileEncSHA256,
					FileSHA256:    urls.FileSHA256,
					FileLength:    proto.Uint64(uint64(size)),
					FileName:      proto.String(fileName),
				},
			}
		})

	case types.MsgAudio:
		convertedPath, err := func() (string, error) {
			if opts.IsVoiceNote {
				return utils.ConvertToOpus(opts.FilePath)
			}
			return utils.ConvertToMP3(opts.FilePath)
		}()
		if err != nil {
			return "", err
		}
		defer os.Remove(convertedPath)
		opts.FilePath = convertedPath
		return sendMedia(opts, whatsmeow.MediaAudio, func(urls uploadedMedia, size int64) *waE2E.Message {
			duration, _ := utils.GetAudioDuration(convertedPath)
			var mimeType string
			var waveform []byte
			if opts.IsVoiceNote {
				mimeType = "audio/ogg; codecs=opus"
				if pcm, err := utils.ReadWaveFile(opts.FilePath); err == nil {
					waveform = utils.GenerateWaveform(pcm, 192)
				}
			} else {
				mimeType = "audio/mpeg"
			}
			return &waE2E.Message{
				AudioMessage: &waE2E.AudioMessage{
					URL:           proto.String(urls.URL),
					DirectPath:    proto.String(urls.DirectPath),
					MediaKey:      urls.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSHA256: urls.FileEncSHA256,
					FileSHA256:    urls.FileSHA256,
					FileLength:    proto.Uint64(uint64(size)),
					PTT:           proto.Bool(opts.IsVoiceNote),
					Seconds:       proto.Uint32(duration),
					Waveform:      waveform,
				},
			}
		})

	case types.MsgSticker:
		tmpWebp := filepath.Join(os.TempDir(), "converted_sticker.webp")
		defer os.Remove(tmpWebp)

		finalPath, err := utils.ToWebp(opts.FilePath, tmpWebp, &types.WebpMetadata{
			Author:     opts.Author,
			PackName:   opts.PackName,
			Categories: opts.Categories,
		})
		if err != nil {
			return "", fmt.Errorf("failed to convert to webp: %v", err)
		}

		opts.FilePath = finalPath

		return sendMedia(opts, whatsmeow.MediaImage, func(urls uploadedMedia, size int64) *waE2E.Message {
			return &waE2E.Message{
				StickerMessage: &waE2E.StickerMessage{
					URL:           proto.String(urls.URL),
					DirectPath:    proto.String(urls.DirectPath),
					MediaKey:      urls.MediaKey,
					Mimetype:      proto.String("image/webp"),
					FileEncSHA256: urls.FileEncSHA256,
					FileSHA256:    urls.FileSHA256,
					FileLength:    proto.Uint64(uint64(size)),
					IsAnimated:    proto.Bool(utils.IsWebpAnimated(finalPath)),
				},
			}
		})

	case types.MsgEdit:
		editMsg := sock.BuildEdit(opts.JID, opts.MessageID, opts.NewMessage)
		_, err := sock.SendMessage(context.Background(), opts.JID, editMsg)
		return "", err
	}

	return "", fmt.Errorf("unknown message type")
}

type uploadedMedia struct {
	URL           string
	DirectPath    string
	MediaKey      []byte
	FileEncSHA256 []byte
	FileSHA256    []byte
}

func sendMedia(opts types.SendOptions, mediaType whatsmeow.MediaType, build func(uploadedMedia, int64) *waE2E.Message) (string, error) {
	data, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	uploaded, err := sock.Upload(context.Background(), data, mediaType)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %v", err)
	}

	fileInfo, err := os.Stat(opts.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %v", err)
	}

	msg := build(uploadedMedia{
		URL:           uploaded.URL,
		DirectPath:    uploaded.DirectPath,
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
	}, fileInfo.Size())

	resp, err := sock.SendMessage(context.Background(), opts.JID, msg)
	if err != nil {
		return "", err
	}

	os.Remove(opts.FilePath)

	return resp.ID, nil
}
