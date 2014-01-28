package soju

import (
	"os"
	"syscall"
	"testing"
	"time"
)

type sojuTest struct {
	StopCalled, StopNowCalled bool
}

func (st *sojuTest) Start() (err error) {
	return
}
func (st *sojuTest) Reconfigure() (err error) {
	return
}
func (st *sojuTest) StopNow(dn DoneNotifier) (err error) {
	st.StopNowCalled = true
	dn.Done()
	return
}
func (st *sojuTest) Stop(dn DoneNotifier) (err error) {
	st.StopCalled = true
	dn.Done()
	return
}

// Gets kill signal,
// Stops and notifies Soju
// returns 0
func TestListenerOnStop(t *testing.T) {
	notificable := new(sojuTest)
	server := new(Server)
	server.SetService(notificable)
	if notificable.StopCalled {
		t.Errorf("Adding a listener to the list should not signal it")
		return
	}
	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- os.Kill
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(1*time.Second, 500*time.Millisecond)
	if result != 0 {
		t.Errorf("return code should be 0 but is [%d] instead", result)
		return
	}
	if !notificable.StopCalled {
		t.Errorf("Stop() method was not called")
		return
	}
	if notificable.StopNowCalled {
		t.Errorf("StopNow() method shouldn't be called.")
		return
	}
}

type firstTimeoutSojuTest struct {
	StopCalled, StopNowCalled bool
}

func (ftst *firstTimeoutSojuTest) Start() (err error) {
	return
}
func (ftst *firstTimeoutSojuTest) Reconfigure() (err error) {
	return
}
func (ftst *firstTimeoutSojuTest) StopNow(dn DoneNotifier) (err error) {
	ftst.StopNowCalled = true
	dn.Done()
	return
}
func (ftst *firstTimeoutSojuTest) Stop(dn DoneNotifier) (err error) {
	ftst.StopCalled = true
	// not calling dn.Done() for testing timeout
	return
}

// Gets kill signal,
// Stops but doesn't notify Soju,
// Gets abort signal
// Stops and notifies Soju
// returns 0
func TestFirstTimeout(t *testing.T) {
	server := new(Server)
	notificable := new(firstTimeoutSojuTest)
	server.SetService(notificable)
	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- os.Kill
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(1*time.Second, 500*time.Millisecond)
	if result != 0 {
		t.Errorf("return code should be 0")
		return
	}
	if !notificable.StopCalled {
		t.Errorf("Stop() method was not called")
		return
	}
	if !notificable.StopNowCalled {
		t.Errorf("StopNow() method was not called.")
		return
	}
}

type secondTimeoutSojuTest struct {
	StopCalled, StopNowCalled bool
}

func (stst *secondTimeoutSojuTest) Start() (err error) {
	return
}
func (stst *secondTimeoutSojuTest) Reconfigure() (err error) {
	return
}
func (stst *secondTimeoutSojuTest) StopNow(dn DoneNotifier) (err error) {
	stst.StopNowCalled = true
	// not calling dn.Done() for testing timeout
	return
}
func (stst *secondTimeoutSojuTest) Stop(dn DoneNotifier) (err error) {
	stst.StopCalled = true
	// not calling dn.Done() for testing timeout
	return
}

// Gets kill signal,
// Stops but doesn't notify Soju,
// Gets abort signal
// Runs stopNow but doesn't notify Soju
// Second timeout
// returns 2
func TestSecondTimeout(t *testing.T) {
	server := new(Server)
	notificable := new(secondTimeoutSojuTest)
	server.SetService(notificable)
	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- os.Kill
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(1*time.Second, 500*time.Millisecond)
	if result != 2 {
		t.Errorf("return code should be 2")
		return
	}
	if !notificable.StopCalled {
		t.Errorf("Stop() method was not called")
		return
	}
	if !notificable.StopNowCalled {
		t.Errorf("StopNow() method was not called.")
		return
	}
}

type sigabrtSojuTest struct {
	StopCalled, StopNowCalled bool
}

func (sst *sigabrtSojuTest) Start() (err error) {
	return
}
func (sst *sigabrtSojuTest) Reconfigure() (err error) {
	return
}
func (sst *sigabrtSojuTest) StopNow(dn DoneNotifier) (err error) {
	sst.StopNowCalled = true
	dn.Done()
	return
}
func (sst *sigabrtSojuTest) Stop(dn DoneNotifier) (err error) {
	sst.StopCalled = true
	dn.Done()
	return
}

// Gets abort signal
// Runs stopNow and notifies Soju
// Stops ok
// returns 0
func TestOnlySIGABRT(t *testing.T) {
	server := new(Server)
	notificable := new(sigabrtSojuTest)
	server.SetService(notificable)
	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- syscall.SIGABRT
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(1*time.Second, 500*time.Millisecond)
	if result != 0 {
		t.Errorf("return code should be 0")
		return
	}
	if notificable.StopCalled {
		t.Errorf("Stop() method shouldn't be called")
		return
	}
	if !notificable.StopNowCalled {
		t.Errorf("StopNow() method was not called.")
		return
	}
}

type sigabrtTimeoutSojuTest struct {
	StopCalled, StopNowCalled bool
}

func (stst *sigabrtTimeoutSojuTest) Start() (err error) {
	return
}
func (stst *sigabrtTimeoutSojuTest) Reconfigure() (err error) {
	return
}
func (stst *sigabrtTimeoutSojuTest) StopNow(dn DoneNotifier) (err error) {
	stst.StopNowCalled = true
	// not calling dn.Done() for testing timeout
	return
}
func (stst *sigabrtTimeoutSojuTest) Stop(dn DoneNotifier) (err error) {
	stst.StopCalled = true
	dn.Done()
	return
}

// Gets abort signal
// Runs stopNow but doesn't notify Soju
// Timeouts
// returns 2
func TestSIGABRTTimeout(t *testing.T) {
	server := new(Server)
	notificable := new(sigabrtTimeoutSojuTest)
	server.SetService(notificable)
	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- syscall.SIGABRT
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(1*time.Second, 500*time.Millisecond)
	if result != 2 {
		t.Errorf("return code should be 2")
		return
	}
	if notificable.StopCalled {
		t.Errorf("Stop() method shouldn't be called")
		return
	}
	if !notificable.StopNowCalled {
		t.Errorf("StopNow() method was not called.")
		return
	}
}

type workerSample struct {
	StopCalled, StopNowCalled bool
}

func (ws *workerSample) StopNow(dn DoneNotifier) (err error) {
	ws.StopNowCalled = true
	RemoveWorker(ws)
	dn.Done()
	return
}
func (ws *workerSample) Stop(dn DoneNotifier) (err error) {
	ws.StopCalled = true
	RemoveWorker(ws)
	dn.Done()
	return
}

// Registers a service and a worker
// Signals the service
// Notifies the worker
// The worker runs it stop method and deregisters from the default soju server
// Everything stops ok (returns 0)
func TestStopWorkers(t *testing.T) {

	notificable := new(sojuTest)
	SetService(notificable)
	if len(defaultSojuServer.workers) != 0 {
		t.Errorf("defaultSojuServer.workers shouldn't have any workers and has %d", len(defaultSojuServer.workers))
		return
	}
	w := new(workerSample)
	AddWorker(w)
	if len(defaultSojuServer.workers) != 1 {
		t.Errorf("defaultSojuServer.workers should have 1 worker and has %d", len(defaultSojuServer.workers))
		return
	}

	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		defaultSojuServer.c <- syscall.SIGKILL
	}()
	// Serve method waits until all gorutines end
	result := Serve(1*time.Second, 500*time.Millisecond)
	if result != 0 {
		t.Errorf("return code should be 2")
		return
	}
	// Test service handlers.
	if !notificable.StopCalled {
		t.Errorf("notificable.Stop() method wasn't called.")
		return
	}
	if notificable.StopNowCalled {
		t.Errorf("notificable.StopNow() method should not be called.")
		return
	}

	// Test worker handlers.
	if !w.StopCalled {
		t.Errorf("w.Stop() method wasn't called.")
		return
	}
	if w.StopNowCalled {
		t.Errorf("w.StopNow() method should not be called.")
		return
	}
	if len(defaultSojuServer.workers) != 0 {
		t.Errorf("defaultSojuServer.workers shouldn't have any workers and has %d. Worker failed to de-register.", len(defaultSojuServer.workers))
		return
	}

}

type nonStoppingWorkerSample struct {
	StopCalled, StopNowCalled bool
}

func (ws *nonStoppingWorkerSample) StopNow(dn DoneNotifier) (err error) {
	ws.StopNowCalled = true
	RemoveWorker(ws)
	dn.Done()
	return
}
func (ws *nonStoppingWorkerSample) Stop(dn DoneNotifier) (err error) {
	ws.StopCalled = true
	// RemoveWorker(ws)
	// dn.Done()
	return
}

// Registers a service and a worker
// Signals the service
// Notifies the worker
// The worker runs it stop method but timeouts
// Soju calls StopNow on both service and worker
// Worker runs StopNow and deregisters
// Everything stops ok (returns 0)
func TestWorkerFirstTimeout(t *testing.T) {

	notificable := new(sojuTest)
	SetService(notificable)
	if len(defaultSojuServer.workers) != 0 {
		t.Errorf("defaultSojuServer.workers shouldn't have any workers and has %d", len(defaultSojuServer.workers))
		return
	}
	w := new(nonStoppingWorkerSample)
	AddWorker(w)
	if len(defaultSojuServer.workers) != 1 {
		t.Errorf("defaultSojuServer.workers should have 1 worker and has %d", len(defaultSojuServer.workers))
		return
	}

	go func() {
		// this goroutine waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		defaultSojuServer.c <- syscall.SIGKILL
	}()
	// Serve method waits until all gorutines end
	result := Serve(1*time.Second, 500*time.Millisecond)
	if result != 0 {
		t.Errorf("return code should be 2")
		return
	}
	// Test service handlers.
	if !notificable.StopCalled {
		t.Errorf("notificable.Stop() method wasn't called.")
		return
	}
	// TODO: known issue: if service stops ok but a worker doesn't, then service's
	// StopNow() handler is called anyways.
	if !notificable.StopNowCalled {
		t.Errorf("notificable.StopNow() method wasn't called.")
		return
	}

	// Test worker handlers.
	if !w.StopCalled {
		t.Errorf("w.Stop() method wasn't called.")
		return
	}
	if !w.StopNowCalled {
		t.Errorf("w.StopNow() method wasn't called.")
		return
	}
	if len(defaultSojuServer.workers) != 0 {
		t.Errorf("defaultSojuServer.workers shouldn't have any workers and has %d. Worker failed to de-register.", len(defaultSojuServer.workers))
		return
	}

}
