package autoffmpeg

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type FFmpegLog struct {
	log *log.Logger
}

func NewFFmpegLog(streamId string) *FFmpegLog {
	pwd, err := os.Getwd()
	if err != nil {
		panic("os.Getwd() error!")
	}
	logFile := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/ffmpeg_log/%s.log", pwd, streamId),
		MaxSize:    128,  // 每个日志文件的最大大小，单位为MB
		MaxBackups: 1,    // 最大保留的日志文件数
		MaxAge:     7,    // 保存的最大天数
		Compress:   true, // 是否压缩旧的日志文件
	}
	logger := log.New(logFile, "", log.LstdFlags)
	return &FFmpegLog{
		log: logger,
	}
}

func (c *FFmpegLog) printf(format string, args ...any) {
	c.log.Printf(format, args...)
}

func (c *FFmpegLog) println(v ...any) {
	c.log.Println(v...)
}
