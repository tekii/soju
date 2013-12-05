package soju

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type DoneNotifier struct {
	called bool // avoid calling Done() more than once
	wg     *sync.WaitGroup
}

func (this *DoneNotifier) Done() {
	if this.called {
		return
	}
	this.called = true
	this.wg.Done()
}

type ShutdownListener interface {
	Shutdown(DoneNotifier)
}

type Server struct {
	shutdownListeners []ShutdownListener
	doneNotifiers     []DoneNotifier
	c                 chan os.Signal
	initialized       bool
	wg                *sync.WaitGroup
	end               chan int
	endCode           int
	timeout           time.Duration
}

func (this *Server) initialize() {
	if this.initialized {
		return
	}
	this.initialized = true
	this.c = make(chan os.Signal, 1)
	this.end = make(chan int, 1)
	signal.Notify(
		this.c,
		syscall.SIGKILL,
		syscall.SIGINT,  //Ctrl + C
		syscall.SIGTERM, //kill command
	)
	this.wg = new(sync.WaitGroup)
	go func() {
		this.wg.Add(1) // count this gorouting as running too
		_ = <-this.c   // wait for os.Signal
		// got signal notify and exit
		for _, listener := range this.shutdownListeners {
			d := DoneNotifier{wg: this.wg}
			this.doneNotifiers = append(this.doneNotifiers, d)
			// Shutdown method must be called in a goroutine.
			// if not the shutdowns are serialized and if one of them hang the whole server hangs.
			go listener.Shutdown(d)
		}
		this.wg.Done() // the function is done (well not really but ...)

		// this channel will be "signaled" when the WaitingGrup reaches 0
		waiter := make(chan int, 1)
		go func() {
			this.wg.Wait()
			waiter <- 0
		}()

		// waits for:
		// 1 - waiter channel been signaled. aka all goroutines are done
		// 2 - waiting time expired
		select {
		case ec := <-waiter:
			// return code 0 (all routines are done)
			this.end <- ec
		case <-time.After(this.timeout):
			// timeout period exceded
			for _, d := range this.doneNotifiers {
				// releases the waiting group using bruteforce
				d.Done()
			}
			// return code 2 => timeout
			this.end <- 2
		}
	}()
}

func (this *Server) AddShutdownListener(listener ShutdownListener) {
	this.initialize()
	this.shutdownListeners = append(this.shutdownListeners, listener)
	this.wg.Add(1)
}

func (this *Server) RemoveShutdownListener(listener ShutdownListener) {
	// TODO: remove listener from list
	// not used now
	this.wg.Add(-1)
}

func (this *Server) Serve(timeout time.Duration) int {
	this.timeout = timeout
	// waits on waitinggroup (forever)
	this.wg.Wait()
	// waits on return code channel
	return <-this.end
}
