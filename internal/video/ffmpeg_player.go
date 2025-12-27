package video

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os/exec"
	"sync"
	"time"
)

// FFmpegPlayer decodes a video file into JPEG frames using external ffmpeg.
//
// This is a pragmatic cross-platform approach:
//   - Works on Windows/macOS/Linux if ffmpeg is available in PATH
//   - Avoids native bindings (GStreamer/DirectShow) complexity
//
// It spawns one ffmpeg process that outputs an MJPEG stream to stdout.
// Audio is not supported.
//
// Thread-safety: Next() is not safe for concurrent calls.
type FFmpegPlayer struct {
	cmd    *exec.Cmd
	stderr bytes.Buffer

	mu     sync.Mutex
	r      *bufio.Reader
	closed bool
}

// Start starts decoding frames from a file.
// fps=0 means "as fast as possible" output cadence (caller should throttle).
func (p *FFmpegPlayer) Start(ctx context.Context, videoPath string, width, height int, fps int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil {
		return fmt.Errorf("player already started")
	}

	args := []string{"-hide_banner", "-loglevel", "error", "-i", videoPath}
	vf := fmt.Sprintf("scale=%d:%d", width, height)
	if fps > 0 {
		vf = vf + fmt.Sprintf(",fps=%d", fps)
	}
	args = append(args, "-vf", vf, "-f", "mjpeg", "pipe:1")

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = &p.stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	p.cmd = cmd
	p.r = bufio.NewReaderSize(stdout, 1<<20)
	p.closed = false
	return nil
}

// Next reads and decodes the next frame.
// Returns (nil, io.EOF) when stream ends.
func (p *FFmpegPlayer) Next() (image.Image, error) {
	p.mu.Lock()
	r := p.r
	p.mu.Unlock()

	if r == nil {
		return nil, fmt.Errorf("player not started")
	}

	var buf bytes.Buffer

	// Find JPEG SOI (0xFFD8)
	for {
		b1, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}
		if b1 != 0xFF {
			continue
		}
		b2, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}
		if b2 == 0xD8 {
			buf.WriteByte(0xFF)
			buf.WriteByte(0xD8)
			break
		}
	}

	// Read until JPEG EOI (0xFFD9)
	prev := byte(0)
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}
		buf.WriteByte(b)
		if prev == 0xFF && b == 0xD9 {
			break
		}
		prev = b
	}

	img, err := jpeg.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (p *FFmpegPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_, _ = p.cmd.Process.Wait()
	}
	p.cmd = nil
	p.r = nil
}

func (p *FFmpegPlayer) Stderr() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stderr.String()
}

// SleepFrame is a helper for basic playback timing.
func SleepFrame(fps int) {
	if fps <= 0 {
		return
	}
	time.Sleep(time.Second / time.Duration(fps))
}
