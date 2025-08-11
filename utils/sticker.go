package utils

import (
	"bot/types"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func ToWebp(inputPath, outputPath string, metadata *types.WebpMetadata) (string, error) {
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute input path: %v", err)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(inputPath), "."))
	videoExts := []string{"mp4", "mov", "mkv", "avi", "webm", "flv", "gif"}

	isVideo := slices.Contains(videoExts, ext)

	var mediaType string
	if isVideo {
		mediaType = "video"
	} else if ext == "webp" {
		mediaType = "webp"
	} else {
		mediaType = "image"
	}

	args := []string{"./node/webp.js", mediaType, absInputPath}

	if metadata != nil {
		if metadata.Author != "" {
			args = append(args, "author", metadata.Author)
		}
		if metadata.PackName != "" {
			args = append(args, "packname", metadata.PackName)
		}
		if len(metadata.Categories) > 0 {
			args = append(args, "categories", strings.Join(metadata.Categories, ","))
		}
	}

	cmd := exec.Command("node", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("node webp conversion failed: %v, output: %s", err, string(output))
	}

	baseName := getBaseName(inputPath)
	tempOutputPath := fmt.Sprintf("%s_sticker.webp", baseName)

	if _, err := os.Stat(tempOutputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("node webp tool did not create expected output file: %s", tempOutputPath)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := copyFile(tempOutputPath, outputPath); err != nil {
		return "", fmt.Errorf("failed to copy output file: %v", err)
	}

	os.Remove(tempOutputPath)

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

func getBaseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i != -1 {
		p = p[i+1:]
	}
	if i := strings.LastIndex(p, "."); i != -1 {
		p = p[:i]
	}
	return p
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}
