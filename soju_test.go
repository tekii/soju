package soju

import (
	"os"
	"testing"
	"time"
)

type sojuTest struct {
	Called bool
}

func (st *sojuTest) Start() (err error) {
	return
}
func (st *sojuTest) Reconfigure() (err error) {
	return
}
func (st *sojuTest) StopNow(dn DoneNotifier) (err error) {
	return
}
func (st *sojuTest) Stop(doneNotifier DoneNotifier) (err error) {
	st.Called = true
	doneNotifier.Done()
	return
}

type brokenSojuTest struct {
	Called bool
}

func (bst *brokenSojuTest) Start() (err error) {
	return
}
func (bst *brokenSojuTest) Reconfigure() (err error) {
	return
}
func (bst *brokenSojuTest) StopNow(dn DoneNotifier) (err error) {
	return
}
func (bst *brokenSojuTest) Stop(doneNotifier DoneNotifier) (err error) {
	bst.Called = true
	// not calling doneNotifier.Done() for testing timeout
	return
}

func TestListenerOnShutdown(t *testing.T) {
	notificable := new(sojuTest)
	server := new(Server)
	server.SetService(notificable)
	if notificable.Called {
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
	result := server.Serve(3*time.Second, 2*time.Second)
	if result != 0 {
		t.Errorf("return code should be 0 but is [%d] instead", result)
		return
	}
	if !notificable.Called {
		t.Errorf("Stop() method was not called")
		return
	}
}

func TestTimeoutFunciona(t *testing.T) {
	server := new(Server)
	notificable := new(brokenSojuTest)
	server.SetService(notificable)
	go func() {
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- os.Kill
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(3*time.Second, 2*time.Second)
	if result == 0 {
		t.Errorf("return code shouldn't be 0")
		return
	}
	if !notificable.Called {
		t.Errorf("Stop() method was not called")
		return
	}
}
