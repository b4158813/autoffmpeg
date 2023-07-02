# autoffmpeg

本项目是ffmpeg自动化执行与信息获取的golang sdk

（灵感来源于直播架构组实习时期开发转码、vqa、预热等服务时，发现总会因为细节考虑不周导致出现难以排查的bug）

## Usage
``` go
    ctx, cancel := context.WithCancel(context.Background())
	params := "ffmpeg xxx"
	param := strings.Split(params, " ")
	cmd := exec.CommandContext(ctx, param[0], param[1:]...)
	autocmd := autoffmpeg.NewFFmpegCmdContext(ctx, cmd)

    // 必须显式异步启动进程
	if err := autocmd.Start(); err != nil {
		fmt.Println("autocmd start error:", err)
		fmt.Println("stderror lines:\n", autocmd.StderrLog())
		return
	}

	fmt.Println("autocmd start success")
	fmt.Println("stderror lines:\n", autocmd.StderrLog())

    // 测试：获取首屏、编码器、档位等信息
	go func() {
		need := true
		for {
			if need && autocmd.IsInitAllDone() {
				fmt.Printf("%+v\n", autocmd.GetAllTransInfo())
				need = false
			}
		}
	}()

    // 测试：结束上下文
	go func() {
		time.Sleep(time.Second * 5)
		cancel()
	}()

    // 必须显式等待进程结束
	if err := autocmd.Wait(); err != nil {
		fmt.Println("autocmd wait error:", err)
		fmt.Println("stderror lines:\n", autocmd.StderrLog())
		return
	}
	fmt.Println("autocmd exec success")
	fmt.Println("stderror lines:\n", autocmd.StderrLog())
```

## Functions

- 能获取ffmpeg转码进程启动后的如下参数：
  - 拉原流的首屏时间
  - 原流的编码器、分辨率档位信息
  - 转码推流的首屏时间
  - 转码推流的编码器、分辨率档位信息
  - 最后若干行标准错误流的异步获取

## Todo