package main

import (
	"autoffmpeg/app"
	"context"
	"os/exec"
)

func main() {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i")
	app.NewFFmpegCmdContext(ctx, cmd)
}
