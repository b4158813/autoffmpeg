package autoffmpeg

type TranscodeInfo struct {
	PullMode  string
	PushMode  string
	PullCodec string
	PushCodec string
	PullRt    int64
	PushRt    int64
}

func NewTranscodeInfo() *TranscodeInfo {
	return transInfo(ModeDefault, ModeDefault, CodecDefault, CodecDefault, 0, 0)
}

func transInfo(pullMode, pushMode, pullCodec, pushCodec string, pullRt, pushRt int64) *TranscodeInfo {
	return &TranscodeInfo{
		PullMode:  pullMode,
		PushMode:  pushMode,
		PullCodec: pullCodec,
		PushCodec: pushCodec,
		PullRt:    pullRt,
		PushRt:    pushRt,
	}
}
