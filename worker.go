package soju

type Worker interface {
	Stop(DoneNotifier) error
	StopNow(DoneNotifier) error
}
