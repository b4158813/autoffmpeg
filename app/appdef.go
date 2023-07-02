package app

import "time"

type sType int

const (
	statusNone sType = iota
	statusInit
	statusRunning
	statusExit
)

const (
	buffReadTimeOutDefault = time.Second * 60
	buffReadLenDefault     = 1024
)
