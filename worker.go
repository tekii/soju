package soju

type Worker interface {
	Stop(DoneNotifier)
	StopNow(DoneNotifier)
}
