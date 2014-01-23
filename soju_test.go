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
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
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
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
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
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
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
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
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
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
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
