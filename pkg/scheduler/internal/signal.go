package internal

type ScheduleSigType int64

const (
	ScheduleSigUnknown ScheduleSigType = 0
	ScheduleSigReady   ScheduleSigType = 1
	ScheduleSigPending ScheduleSigType = 2
)

type ScheduleSignal interface {
	getType() ScheduleSigType
}
