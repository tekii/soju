package soju

import (
	"os"
	"testing"
	"time"
)

type sojuTest struct {
	Called bool
}

func (st *sojuTest) Shutdown(doneNotifier DoneNotifier) {
	st.Called = true
	doneNotifier.Done()
}

type brokenSojuTest struct {
	Called bool
}

func (bst *brokenSojuTest) Shutdown(doneNotifier DoneNotifier) {
	bst.Called = true
	// not calling doneNotifier.Done() for testing timeout
}

func TestListenerOnShutdown(t *testing.T) {
	notificable := new(sojuTest)
	server := new(Server)
	server.AddShutdownListener(notificable)
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
	result := server.Serve(3 * time.Second)
	if result != 0 {
		t.Errorf("return code should be 0 but is [%d] instead", result)
		return
	}
	if !notificable.Called {
		t.Errorf("Shutdown() method was not called")
		return
	}
}

func TestTimeoutFunciona(t *testing.T) {
	server := new(Server)
	notificable := new(brokenSojuTest)
	server.AddShutdownListener(notificable)
	go func() {
		// this goruting waits 1/2 second and then signal the channel pretending an external signal
		time.Sleep(time.Millisecond * 500)
		// this can be done here because test is in the same package but this channel is private
		server.c <- os.Kill
	}()
	// Serve method waits until all gorutines end
	result := server.Serve(3 * time.Second)
	if result == 0 {
		t.Errorf("return code shouldn't be 0")
		return
	}
	if !notificable.Called {
		t.Errorf("Shutdown() method was not called")
		return
	}
}
