package autoffmpeg

const (
	CatchStderrmaxSizeDefault = 3
)

type CatchStderr struct {
	has          bool
	maxSize      int
	stderrLineLs []string
	stderrLine   string
}

func NewCatchStderr(has bool, maxSize int) *CatchStderr {
	return &CatchStderr{
		has:          has,
		maxSize:      maxSize,
		stderrLineLs: make([]string, 0),
		stderrLine:   "",
	}
}

func (c *CatchStderr) catchLine(line string) {
	// TODO 原子性（暂不需要）
	// 获取最后LAST_STDERR_LINE_CNT行错误流数据用于上报
	c.stderrLineLs = append(c.stderrLineLs, line)
	if len(c.stderrLineLs) > c.maxSize {
		c.stderrLineLs = (c.stderrLineLs)[1:]
	}
}

// 需要等待发现错误流中的信息
type NeedToDo struct {
	NeedPullFSCR *singleToDo
	NeedPushCR   *singleToDo
	NeedPushFS   *singleToDo
}

func NewNeedToDo(isPullFSCR, isPushCR, isPushFS bool) *NeedToDo {
	ret := &NeedToDo{}
	if isPullFSCR {
		ret.NeedPullFSCR = newSingleToDo()
	}
	if isPushCR {
		ret.NeedPushCR = newSingleToDo()
	}
	if isPushFS {
		ret.NeedPushFS = newSingleToDo()
	}
	return ret
}
