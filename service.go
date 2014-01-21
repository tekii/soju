package soju

type Service interface {
	Reconfigure() error
	Start() error
	Stop(DoneNotifier) error
	StopNow(DoneNotifier) error
}
