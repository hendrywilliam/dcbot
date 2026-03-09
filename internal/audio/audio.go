package audio

import (
	"context"
	"fmt"
	"os/exec"
)

type Audio struct{}

func (a *Audio) Encode(ctx context.Context, name string, data chan<- []byte, done chan bool) error {
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-i", fmt.Sprintf("./media/%s", name),
		"-ac", "2",
		"-ar", "48000",
		"-c:a", "libopus",
		"-f", "opus",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	buf := make([]byte, 3840)
	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			return nil
		default:
			n, err := stdout.Read(buf)
			if err != nil {
				cmd.Wait()
				done <- true
				return nil
			}
			if n > 0 {
				frame := make([]byte, n)
				copy(frame, buf[:n])
				data <- frame
			}
		}
	}
}
