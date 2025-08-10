package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-audio/wav"
)

func checkFFmpeg(cmd string) {
	if _, err := exec.LookPath(cmd); err != nil {
		panic(fmt.Sprintf("%s not installed or not in PATH", cmd))
	}
}

func convertAudio(input, codec, bitrate, ext string) (string, error) {
	checkFFmpeg("ffmpeg")
	out := strings.TrimSuffix(input, filepath.Ext(input)) + ext
	cmd := exec.Command("ffmpeg", "-y", "-i", input, "-c:a", codec, "-b:a", bitrate, out)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %v - %s", err, output)
	}
	return out, nil
}

func ConvertToOpus(input string) (string, error) {
	return convertAudio(input, "libopus", "128k", ".opus.ogg")
}

func ConvertToMP3(input string) (string, error) {
	return convertAudio(input, "libmp3lame", "192k", ".converted.mp3")
}

func GetAudioDuration(filePath string) (uint32, error) {
	checkFFmpeg("ffprobe")
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe error: %v", err)
	}
	var res struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		return 0, fmt.Errorf("parse ffprobe output: %v", err)
	}
	secs, err := strconv.ParseFloat(res.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %v", err)
	}
	return uint32(secs + 0.5), nil
}

func ReadWaveFile(filePath string) ([]int, error) {
	if filepath.Ext(filePath) == ".wav" {
		if s, err := readWavDirect(filePath); err == nil {
			return s, nil
		}
	}
	return readPcmViaFfmpeg(filePath)
}

func readWavDirect(filePath string) ([]int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file")
	}
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Data, nil
}

func readPcmViaFfmpeg(filePath string) ([]int, error) {
	checkFFmpeg("ffmpeg")
	cmd := exec.Command("ffmpeg", "-i", filePath, "-f", "s16le", "-acodec", "pcm_s16le", "-ac", "1", "-ar", "16000", "-hide_banner", "-loglevel", "quiet", "pipe:1")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg PCM extraction failed: %v", err)
	}
	data := out.Bytes()
	samples := make([]int, len(data)/2)
	for i := range samples {
		samples[i] = int(int16(binary.LittleEndian.Uint16(data[i*2:])))
	}
	return samples, nil
}

func GenerateWaveform(samples []int, count int) []byte {
	if len(samples) == 0 {
		return nil
	}
	if count > len(samples) {
		count = len(samples)
	}
	step := len(samples) / count
	waveform := make([]byte, count)
	for i := 0; i < count; i++ {
		sum := 0
		for j := 0; j < step; j++ {
			idx := i*step + j
			if idx >= len(samples) {
				break
			}
			sum += samples[idx] * samples[idx]
		}
		rms := int(math.Sqrt(float64(sum) / float64(step)))
		if rms > 32767 {
			rms = 32767
		}
		waveform[i] = byte((rms * 100) / 32767)
	}
	return waveform
}
