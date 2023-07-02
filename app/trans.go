package app

import "autoffmpeg/utils"

type TrancodeInfo struct {
	PullMode  string
	PushMode  string
	PullCodec string
	PushCodec string
	PullRt    int64
	PushRt    int64
}

func NewTranscodeInfo() *TrancodeInfo {
	return transInfo(utils.ModeDefault, utils.ModeDefault, utils.CodecDefault, utils.CodecDefault, 0, 0)
}

func transInfo(pullMode, pushMode, pullCodec, pushCodec string, pullRt, pushRt int64) *TrancodeInfo {
	return &TrancodeInfo{
		PullMode:  pullMode,
		PushMode:  pushMode,
		PullCodec: pullCodec,
		PushCodec: pushCodec,
		PullRt:    pullRt,
		PushRt:    pushRt,
	}
}
