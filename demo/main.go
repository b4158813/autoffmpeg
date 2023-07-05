package main

import (
	autoffmpeg "autoffmpeg/app"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	params := "ffmpeg -y -stream_loop -1 -i wlxtest.mp4 -c:v hevc_videotoolbox -c:a aac -b:a 128k -ar 44100 -f flv output"
	// params := "ffmpeg -i"
	param := strings.Split(params, " ")
	cmd := exec.CommandContext(ctx, param[0], param[1:]...)
	autocmd := autoffmpeg.NewFFmpegCmdContext(ctx, cmd)
	defer fmt.Println("stderror lines:\n", autocmd.StderrLog())
	fmt.Println("autocmd init done")
	if err := autocmd.Start(); err != nil {
		fmt.Println("autocmd start error:", err)
		return
	}

	fmt.Println("autocmd start success")

	go func() {
		select {
		case <-autocmd.InitAllDoneSig():
			fmt.Printf("%+v\n", autocmd.GetAllTransInfo())
		}
	}()

	go func(cancel func()) {
		time.Sleep(time.Second * 5)
		cancel()
	}(cancel)

	if err := autocmd.Wait(); err != nil {
		fmt.Println("autocmd wait error:", err)
		return
	}
	fmt.Println("autocmd exec success")
}
