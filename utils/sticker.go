package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ToWebp(inputPath, outputPath string) (string, bool, error) {
	ext := strings.ToLower(strings.TrimPrefix(getFileExt(inputPath), "."))
	tmpDir := "./webp_tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", false, err
	}

	localWebp := fmt.Sprintf("%s/%s.webp", tmpDir, getBaseName(inputPath))
	var cmd *exec.Cmd

	if isVideo(ext) {
		cmd = exec.Command("ffmpeg", "-y", "-i", inputPath,
			"-t", "8",
			"-vf", "scale=512:512:force_original_aspect_ratio=decrease,fps=15",
			"-c:v", "libwebp",
			"-lossless", "0",
			"-q:v", "50",
			"-loop", "0",
			"-an",
			"-preset", "picture",
			localWebp)
	} else {
		cmd = exec.Command("cwebp", "-q", "80", inputPath, "-o", localWebp)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", false, fmt.Errorf("conversion failed: %v, %s", err, string(out))
	}

	if fi, _ := os.Stat(localWebp); fi != nil && fi.Size() > 800*1024 {
		if isVideo(ext) {
			cmd = exec.Command("ffmpeg", "-y", "-i", localWebp,
				"-c:v", "libwebp", "-lossless", "0", "-q:v", "60", "-preset", "picture", localWebp)
		} else {
			cmd = exec.Command("cwebp", "-q", "60", inputPath, "-o", localWebp)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", false, fmt.Errorf("recompression failed: %v, %s", err, string(out))
		}
	}

	if err := copyFile(localWebp, outputPath); err != nil {
		return "", false, err
	}

	return outputPath, IsWebpAnimated(outputPath), nil
}

func isVideo(ext string) bool {
	for _, v := range []string{"mp4", "mov", "mkv", "avi", "webm", "flv", "gif"} {
		if ext == v {
			return true
		}
	}
	return false
}

func getFileExt(p string) string {
	if i := strings.LastIndex(p, "."); i != -1 {
		return p[i:]
	}
	return ""
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
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

func IsWebpAnimated(path string) bool {
	out, err := exec.Command("webpmux", "-info", path).CombinedOutput()
	if err != nil {
		return false
	}
	s := string(out)
	return strings.Contains(s, "Number of frames") && !strings.Contains(s, "Number of frames: 1")
}
