package main

import (
	autoffmpeg "autoffmpeg/app"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	ctx := context.Background()
	params := "ffmpeg -y -stream_loop -1 -i wlxtest.mp4 -c:v hevc_videotoolbox -c:a aac -b:a 128k -ar 44100 -f flv output"
	// params := "ffmpeg -i"
	param := strings.Split(params, " ")
	cmd := exec.CommandContext(ctx, param[0], param[1:]...)
	autocmd := autoffmpeg.NewFFmpegCmdContext(ctx, cmd)
	fmt.Println("autocmd init done")
	if err := autocmd.Start(); err != nil {
		fmt.Println("autocmd start error:", err)
		fmt.Println("stderror lines:\n", autocmd.StderrLog())
		return
	}

	fmt.Println("autocmd start success")
	fmt.Println("stderror lines:\n", autocmd.StderrLog())

	go func() {
		need := true
		for {
			if need && autocmd.IsInitAllDone() {
				fmt.Printf("%+v\n", autocmd.GetAllTransInfo())
				need = false
			}
		}
	}()

	if err := autocmd.Wait(); err != nil {
		fmt.Println("autocmd wait error:", err)
		fmt.Println("stderror lines:\n", autocmd.StderrLog())
		return
	}
	fmt.Println("autocmd exec success")
	fmt.Println("stderror lines:\n", autocmd.StderrLog())
}
