package media

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ConvertAudioToOgg(media []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-i", "-", "-c:a", "libopus", "-f", "ogg", "-")
	var outStdout bytes.Buffer
	var outStderr bytes.Buffer
	cmd.Stdout = &outStdout
	cmd.Stderr = &outStderr
	cmd.Stdin = bytes.NewReader(media)

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %v\nstderr: %s", err, outStderr.String())
	}

	return outStdout.Bytes(), nil

}
