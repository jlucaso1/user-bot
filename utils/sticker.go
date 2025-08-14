package utils

import (
	"bot/types"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func ToWebp(inputPath, outputPath string, metadata *types.WebpMetadata) (string, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute input path: %v", err)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(inputPath), "."))
	videoExts := []string{"mp4", "mov", "mkv", "avi", "webm", "flv", "gif"}
	isAnimated := slices.Contains(videoExts, ext)

	tempFFmpegOut := outputPath + ".ffmpeg.webp"
	defer os.Remove(tempFFmpegOut)

	if err := convertWithFFmpeg(absInputPath, tempFFmpegOut, isAnimated); err != nil {
		return "", err
	}

	if metadata != nil && (metadata.Author != "" || metadata.PackName != "") {
		return addExifToWebp(tempFFmpegOut, outputPath, metadata)
	}

	if err := os.Rename(tempFFmpegOut, outputPath); err != nil {
		return "", fmt.Errorf("failed to move temp webp to final path: %w", err)
	}

	return outputPath, nil
}

func convertWithFFmpeg(inputPath, outputPath string, isAnimated bool) error {
	var cmd *exec.Cmd

	if isAnimated {
		cmd = exec.Command("ffmpeg", "-y", "-i", inputPath,
			"-vcodec", "libwebp",
			"-vf", "fps=15,scale=512:512:flags=lanczos:force_original_aspect_ratio=decrease",
			"-loop", "0",
			"-preset", "default",
			"-an",
			"-fps_mode", "vfr",
			"-s", "512:512",
			"-f", "webp",
			outputPath,
		)
	} else {
		cmd = exec.Command("ffmpeg", "-y", "-i", inputPath,
			"-vcodec", "libwebp",
			"-vf", "scale=512:512:flags=lanczos:force_original_aspect_ratio=decrease",
			"-lossless", "1",
			"-preset", "default",
			"-an",
			"-fps_mode", "vfr",
			"-s", "512:512",
			"-f", "webp",
			outputPath,
		)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %s\n%v", string(output), err)
	}
	return nil
}

func addExifToWebp(inputPath, outputPath string, metadata *types.WebpMetadata) (string, error) {
	jsonPayload := map[string]interface{}{
		"sticker-pack-id":        "https://github.com/AstroX11/user-bot",
		"sticker-pack-name":      metadata.PackName,
		"sticker-pack-publisher": metadata.Author,
		"emojis":                 metadata.Categories,
	}
	if len(metadata.Categories) == 0 {
		jsonPayload["emojis"] = []string{""}
	}

	jsonBytes, err := json.Marshal(jsonPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal exif json: %w", err)
	}

	exifHeader := []byte{
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x41, 0x57, 0x07, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
	}

	exifData := append(exifHeader, jsonBytes...)
	binary.LittleEndian.PutUint32(exifData[14:], uint32(len(jsonBytes)))

	exifFilePath := inputPath + ".exif"
	if err := os.WriteFile(exifFilePath, exifData, 0644); err != nil {
		return "", fmt.Errorf("failed to write exif data to file: %w", err)
	}
	defer os.Remove(exifFilePath)

	cmd := exec.Command("webpmux", "-set", "exif", exifFilePath, inputPath, "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("webpmux failed: %s\n%v", string(output), err)
	}

	return outputPath, nil
}

func IsWebpAnimated(path string) bool {
	out, err := exec.Command("webpmux", "-info", path).CombinedOutput()
	if err != nil {
		return false
	}
	s := string(out)
	return strings.Contains(s, "Number of frames") && !strings.Contains(s, "Number of frames: 1")
}
