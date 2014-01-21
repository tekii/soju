package soju

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type DoneNotifier interface {
	Done()
}

type DefaultDoneNotifier struct {
	called bool // avoid calling Done() more than once
	wg     *sync.WaitGroup
	once   sync.Once
}

func (dn *DefaultDoneNotifier) Done() {
	if dn.called {
		return
	}
	dn.called = true
	dn.once.Do(dn.wg.Done)
	return
}

type Server struct {
	service        Service
	workers        []Worker
	doneNotifiers  []DoneNotifier
	c              chan os.Signal
	initialized    bool
	wg             *sync.WaitGroup
	end            chan int
	endCode        int
	stopTimeout    time.Duration
	stopNowTimeout time.Duration
}

func (s *Server) SetService(service Service) {
	s.service = service
	return
}

func (s *Server) initialize() {
	if s.initialized {
		return
	}
	s.initialized = true
	s.c = make(chan os.Signal, 1)
	s.end = make(chan int, 1)
	signal.Notify(
		s.c,
		syscall.SIGKILL,
		syscall.SIGINT,  //Ctrl + C
		syscall.SIGTERM, //kill command
	)
	s.wg = new(sync.WaitGroup)
	go func() {
		s.wg.Add(1) // count this gorouting as running too
		_ = <-s.c   // wait for os.Signal
		// got signal notify and exit
		d := DefaultDoneNotifier{wg: s.wg}
		d.wg.Add(1)
		s.doneNotifiers = append(s.doneNotifiers, &d)
		//Stop the service asynchronously so that it doesn't hang the server
		go s.service.Stop(&d)
		s.NotifyWorkers()

		s.wg.Done() // the function is done (well not really but ...)

		// this channel will be "signaled" when the WaitingGrup reaches 0
		waiter := make(chan int, 1)
		go func() {
			s.wg.Wait()
			waiter <- 0
		}()

		// waits for:
		// 1 - waiter channel been signaled. aka all goroutines are done
		// 2 - waiting time expired
		select {
		case ec := <-waiter:
			// return code 0 (all routines are done)
			s.end <- ec
		case <-time.After(s.stopTimeout):
			// timeout period exceded
			for _, d := range s.doneNotifiers {
				// releases the waiting group using bruteforce
				d.Done()
			}
			// return code 2 => timeout
			s.end <- 2
		}
	}()
}

func (s *Server) AddWorker(worker Worker) {
	s.workers = append(s.workers, worker)
	s.wg.Add(1)
}

func (s *Server) RemoveWorker(worker Worker) {
	// TODO: remove worker from list
	// not used now
	s.wg.Add(-1)
}

func (s *Server) NotifyWorkers() {
	for _, worker := range s.workers {
		d := DefaultDoneNotifier{wg: s.wg}
		d.wg.Add(1)
		s.doneNotifiers = append(s.doneNotifiers, &d)
		// Stop method must be called in a goroutine.
		// if not the shutdowns are serialized and if one of them hang the whole server hangs.
		go worker.Stop(&d)
	}
}

func (s *Server) Serve(stopTimeout, stopNowTimeout time.Duration) int {
	s.stopTimeout = stopTimeout
	s.stopNowTimeout = stopNowTimeout
	s.initialize()
	// waits on waitinggroup (forever)
	s.wg.Wait()
	// waits on return code channel
	return <-s.end
}
