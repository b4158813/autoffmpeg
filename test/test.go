package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	params := "ffmpeg -y -stream_loop -1 -i wlxtest.mp4 -c:v hevc_videotoolbox -c:a aac -b:a 128k -ar 44100 -f flv output"
	// params := "ffmpeg -i"
	param := strings.Split(params, " ")
	exe := exec.CommandContext(ctx, param[0], param[1:]...)

	fmt.Println(exe.Process)

	if err := exe.Start(); err != nil {
		fmt.Println("exe.start failed, error:", err)
		// fmt.Println("stderror lines:", strings.Join(last_stderr_lines, "\n"))
		return
	}

	fmt.Println("exe.start success")
	fmt.Println(exe.Process)

	time.Sleep(time.Second * 5)
	if err := exe.Wait(); err != nil {
		fmt.Println("exe.wait failed, error:", err)
	}
	fmt.Println(exe.Process)
	// fmt.Println("stderror lines:", strings.Join(last_stderr_lines, "\n"))
}
