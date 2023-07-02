package main

import (
	autoffmpeg "autoffmpeg/app"
	"context"
	"os/exec"
)

func main() {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i")
	autoffmpeg.NewFFmpegCmdContext(ctx, cmd)
}
