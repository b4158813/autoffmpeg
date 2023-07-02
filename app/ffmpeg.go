package app

import (
	"autoffmpeg/utils"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"
)

type ffmpegCmdMethods interface {
	setStatus(sType)
	checkInitAllDone() bool

	ReadFromBuff()
	SetBuffTimeOut(time.Duration)
	ReadString() string
	ResetBuff()
	ReadStringWithTimeOut() string

	StartRaw() error
	StartWhileGetAllInfo() error
	WaitInitAllDone() error
}

type ffmpegCmdPullMethods interface {
	GetPullCodec() string
}

type ffmpegCmdPushMethods interface {
	GetPushCodec() string
}

type ffmpegCmdTranscodeMethods interface {
	GetPullCodec() string
	GetPushCodec() string
	GetPullFirstScreenMilli() int64
	GetPushFirstScreenMilli() int64
}

type FFmpegCmd struct {
	Cmd  *exec.Cmd
	ctx  context.Context
	buff []byte
	pr   *io.PipeReader
	pw   *io.PipeWriter
	// buff read time out seconds
	buffReadTimeout time.Duration
	status          sType

	isInfoAllDone   bool
	allInfoDoneChan chan struct{}
	transInfo       *TrancodeInfo
	needToDo        *NeedToDo

	stTimeUnixMilli int64
	catchStderr     *CatchStderr
}

var _ ffmpegCmdMethods = &FFmpegCmd{}
var _ ffmpegCmdPullMethods = &FFmpegCmd{}
var _ ffmpegCmdPushMethods = &FFmpegCmd{}
var _ ffmpegCmdTranscodeMethods = &FFmpegCmd{}

func NewFFmpegCmd(cmd *exec.Cmd) *FFmpegCmd {
	if !utils.ContainsFFmpeg(cmd.Path) {
		panic("cmd path ffmpeg not include")
	}
	pr, pw := io.Pipe()
	cmd.Stderr = pw
	ret := &FFmpegCmd{
		Cmd:             cmd,
		ctx:             nil,
		buff:            make([]byte, buffReadLenDefault),
		pr:              pr,
		pw:              pw,
		buffReadTimeout: buffReadTimeOutDefault,
		status:          statusNone,
		allInfoDoneChan: make(chan struct{}),
		transInfo:       NewTranscodeInfo(),
		needToDo:        NewNeedToDo(true, true, true),
		stTimeUnixMilli: 0,
		catchStderr:     NewCatchStderr(true, CatchStderrmaxSizeDefault),
	}
	return ret
}

func NewFFmpegCmdContext(ctx context.Context, cmd *exec.Cmd) *FFmpegCmd {
	if ctx == nil {
		panic("nil Context")
	}
	ffmpegCmd := NewFFmpegCmd(cmd)
	ffmpegCmd.ctx = ctx
	return ffmpegCmd
}

func (c *FFmpegCmd) setStatus(t sType) {
	c.status = t
}

// TODO 不够抽象
func (c *FFmpegCmd) checkInitAllDone() bool {
	if c.isInfoAllDone {
		return true
	}
	temp := []bool{
		c.needToDo.NeedPullFSCR.HasAndIsDone(),
		c.needToDo.NeedPushCR.HasAndIsDone(),
		c.needToDo.NeedPushFS.HasAndIsDone(),
	}
	if utils.ReduceBoolListAllTrue(temp) {
		c.isInfoAllDone = true
		return true
	}
	return false
}

func (c *FFmpegCmd) ReadFromBuff() {
	c.pr.Read(c.buff)
}

func (c *FFmpegCmd) SetBuffTimeOut(seconds time.Duration) {
	c.buffReadTimeout = seconds
}

func (c *FFmpegCmd) ReadString() string {
	c.ReadFromBuff()
	return string(c.buff)
}

func (c *FFmpegCmd) ResetBuff() {
	c.buff = make([]byte, len(c.buff))
}

func (c *FFmpegCmd) ReadStringWithTimeOut() string {
	sig := make(chan struct{})
	go func() {
		c.ReadFromBuff()
		sig <- struct{}{}
	}()
	timer := time.NewTimer(c.buffReadTimeout)
	for {
		select {
		case <-c.ctx.Done():
			return ""
		case <-timer.C:
			return ""
		case <-sig:
			return string(c.buff)
		}
	}
}

func (c *FFmpegCmd) GetPullCodec() (codec string) {
	codec = c.transInfo.PullCodec
	return
}

func (c *FFmpegCmd) GetPushCodec() (codec string) {
	codec = c.transInfo.PushCodec
	return
}

func (c *FFmpegCmd) GetPullFirstScreenMilli() (rt int64) {
	rt = c.transInfo.PullRt
	return
}

func (c *FFmpegCmd) GetPushFirstScreenMilli() (rt int64) {
	rt = c.transInfo.PushRt
	return
}

func (c *FFmpegCmd) StartRaw() error {
	if c.status != statusNone {
		return errors.New("FFmpegCmd status is not none")
	}
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	c.setStatus(statusInit)
	return nil
}

func (c *FFmpegCmd) StartWhileGetAllInfo() error {
	if err := c.StartRaw(); err != nil {
		return err
	}
	go c.DoAfterStart()
	return nil
}

func (c *FFmpegCmd) WaitInitAllDone() error {
	<-c.allInfoDoneChan
	return nil
}

// block
func (c *FFmpegCmd) DoAfterStart() error {
	c.stTimeUnixMilli = time.Now().UnixMilli()
	for {
		select {
		case <-c.ctx.Done():
			return errors.New("ctx done, quit ffmpeg cmd stderr monitor")
		default:
			this_line := c.ReadStringWithTimeOut()
			if len(strings.TrimSpace(this_line)) == 0 {
				break
			}
			this_line_lower := strings.ToLower(this_line)

			if c.catchStderr.has {
				c.catchStderr.catchLine(this_line)
			}

			c.StderrGetAllInfo(this_line_lower)

			c.ResetBuff()
		}
	}
}

func (c *FFmpegCmd) StderrGetAllInfo(line string) {
	// 转码流程：拉源流 -> 转推转码流
	if c.needToDo.NeedPullFSCR.HasAndNotDone() {
		// 拉到源流后，stderr会有Video信息，且这一行一定会包含codec和分辨率信息
		if utils.ContainsVideoInfoAll(line) {
			c.transInfo.PullRt = time.Now().UnixMilli() - c.stTimeUnixMilli
			c.transInfo.PullMode = utils.GetStdModeFromString(line)
			c.transInfo.PullCodec = utils.GetStdCodecFromString(line)

			c.needToDo.NeedPullFSCR.MakeDone()
		}
	} else if c.needToDo.NeedPushCR.HasAndNotDone() {
		// 转码成功后，stderr会有Video信息，且这一行一定会包含codec和分辨率信息
		if utils.ContainsVideoInfoAll(line) {
			c.transInfo.PushMode = utils.GetStdModeFromString(line)
			c.transInfo.PushCodec = utils.GetStdCodecFromString(line)

			c.needToDo.NeedPushCR.MakeDone()
		}
	} else if c.needToDo.NeedPushFS.HasAndNotDone() {
		// 转推成功后，stderr会不停地打印如下line
		// frame=  332 fps= 35 q=36.0 size=    2048kB time=00:00:15.13 bitrate=1108.6kbits/s dup=0 drop=17 speed= 1.6x    4kB time=00:00:01.80 bitrate= 335.7kbits/s dup=0 drop=1 speed=3.48x
		if utils.ContainsTranscodePushInfoAll(line) {
			c.transInfo.PushRt = time.Now().UnixMilli() - c.stTimeUnixMilli

			c.needToDo.NeedPushFS.MakeDone()
		}
	}

	if !c.isInfoAllDone && c.checkInitAllDone() {
		go func() {
			c.allInfoDoneChan <- struct{}{}
		}()
	}
}
