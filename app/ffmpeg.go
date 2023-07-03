package autoffmpeg

import (
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

	IsInitAllDone() bool
	InitAllDoneSig() <-chan struct{}

	ReadFromBuff()
	SetBuffTimeOut(time.Duration)
	ReadString() string
	ResetBuff()
	ReadStringWithTimeOut() string

	StartRaw() error
	Start() error
	WaitRaw() error
	Wait() error
	WaitWithInitInfo()

	StderrGetAllInfo(string)
	StderrLog() string
}

type ffmpegCmdTranscodeMethods interface {
	GetAllTransInfo() TranscodeInfo
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

	isInitAllDone  bool
	initAllDoneSig chan struct{}
	transInfo      *TranscodeInfo
	needToDo       *NeedToDo

	stTimeUnixMilli int64
	catchStderr     *CatchStderr
}

var _ ffmpegCmdMethods = &FFmpegCmd{}
var _ ffmpegCmdTranscodeMethods = &FFmpegCmd{}

func NewFFmpegCmd(cmd *exec.Cmd) *FFmpegCmd {
	if !ContainsFFmpeg(cmd.Path) {
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
		isInitAllDone:   false,
		initAllDoneSig:  make(chan struct{}),
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
	if c.isInitAllDone {
		return true
	}
	temp := []bool{
		c.needToDo.NeedPullFSCR.HasAndIsDone(),
		c.needToDo.NeedPushCR.HasAndIsDone(),
		c.needToDo.NeedPushFS.HasAndIsDone(),
	}
	if ReduceBoolListAllTrue(temp) {
		c.isInitAllDone = true
		go func() {
			c.initAllDoneSig <- struct{}{}
		}()
		return true
	}
	return false
}

func (c *FFmpegCmd) IsInitAllDone() bool {
	return c.checkInitAllDone()
}

func (c *FFmpegCmd) InitAllDoneSig() <-chan struct{} {
	return c.initAllDoneSig
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

func (c *FFmpegCmd) Start() error {
	if err := c.StartRaw(); err != nil {
		return err
	}
	go c.DoAfterStart()
	return nil
}

func (c *FFmpegCmd) WaitRaw() error {
	return c.Cmd.Wait()
}

func (c *FFmpegCmd) Wait() error {
	go c.WaitWithInitInfo()
	return c.Cmd.Wait()
}

func (c *FFmpegCmd) WaitWithInitInfo() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if c.isInitAllDone {
				return
			}
			if c.Cmd.ProcessState != nil {
				return
			}
		}
	}
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
		if ContainsVideoInfoAll(line) {
			c.transInfo.PullRt = time.Now().UnixMilli() - c.stTimeUnixMilli
			c.transInfo.PullMode = GetStdModeFromString(line)
			c.transInfo.PullCodec = GetStdCodecFromString(line)

			c.needToDo.NeedPullFSCR.MakeDone()
		}
	} else if c.needToDo.NeedPushCR.HasAndNotDone() {
		// 转码成功后，stderr会有Video信息，且这一行一定会包含codec和分辨率信息
		if ContainsVideoInfoAll(line) {
			c.transInfo.PushMode = GetStdModeFromString(line)
			c.transInfo.PushCodec = GetStdCodecFromString(line)

			c.needToDo.NeedPushCR.MakeDone()
		}
	} else if c.needToDo.NeedPushFS.HasAndNotDone() {
		// 转推成功后，stderr会不停地打印如下line
		// frame=  332 fps= 35 q=36.0 size=    2048kB time=00:00:15.13 bitrate=1108.6kbits/s dup=0 drop=17 speed= 1.6x    4kB time=00:00:01.80 bitrate= 335.7kbits/s dup=0 drop=1 speed=3.48x
		if ContainsTranscodePushInfoAll(line) {
			c.transInfo.PushRt = time.Now().UnixMilli() - c.stTimeUnixMilli

			c.needToDo.NeedPushFS.MakeDone()
		}
	}

	c.checkInitAllDone()
}

func (c *FFmpegCmd) GetAllTransInfo() TranscodeInfo {
	return *c.transInfo
}

func (c *FFmpegCmd) StderrLog() string {
	return c.catchStderr.stderrLog()
}
