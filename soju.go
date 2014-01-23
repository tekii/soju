package soju

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// DoneNotifiers are passed to signal handlers so that the service and workers
// are able to inform when they have correctly finished.
type DoneNotifier interface {
	Done()
}

// Default implementation of DoneNotifier
type DefaultDoneNotifier struct {
	wg   *sync.WaitGroup
	once sync.Once
}

func (dn *DefaultDoneNotifier) Done() {
	dn.once.Do(dn.wg.Done)
	return
}

// A Soju Server receives the OS signals and notifies it's service and all the registered
// workers.
type Server struct {
	// Main service
	service Service

	// Workers
	workers []Worker

	// Signals notifiers
	doneNotifiers []DoneNotifier
	wg            *sync.WaitGroup

	// SIGABRT notifiers
	stopNowNotifiers []DoneNotifier
	stopNowWG        *sync.WaitGroup

	c           chan os.Signal
	initialized sync.Once
	end         chan int
	endCode     int

	// Timeouts
	stopTimeout    time.Duration
	stopNowTimeout time.Duration
}

// Sets the server's managed service.
func (s *Server) SetService(service Service) {
	s.service = service
	return
}

// It also notifies, maybe notifyThenWaitOrTimeout?
func (s *Server) waitOrTimeout(sig os.Signal, wg *sync.WaitGroup, timeout time.Duration) {

	// Notify signal
	s.notify(sig, s.service)
	// If any, notify all registered workers.
	s.NotifyWorkers(sig)

	// This channel will be "signaled" when the WaitingGrup reaches 0
	waiter := make(chan int, 1)
	go func() {
		wg.Wait()
		waiter <- 0
	}()

	// Waits for:
	select {
	// 1 - Waiter channel to be signaled. AKA all goroutines are done
	case ec := <-waiter:
		// TODO: clear s.doneNotifiers and s.wg?
		s.end <- ec
	// 2 - Timeout
	case <-time.After(timeout):

		// No more wait... stop everything now!
		if sig == syscall.SIGABRT {

			// Releases the waiting groups using bruteforce
			for i := range s.doneNotifiers {
				s.doneNotifiers[i].Done()
			}
			for i := range s.stopNowNotifiers {
				s.stopNowNotifiers[i].Done()
			}
			// return code 2 => timeout
			s.end <- 2

			return

		}

		// Send SIGABRT signal.
		s.waitOrTimeout(syscall.SIGABRT, s.stopNowWG, s.stopNowTimeout)

	}

	return

}

func (s *Server) initialize() {

	s.c = make(chan os.Signal, 1)
	s.end = make(chan int, 1)

	signal.Notify(
		s.c,
		syscall.SIGKILL,
		syscall.SIGINT,  // Ctrl + C
		syscall.SIGTERM, // kill command
		syscall.SIGABRT, // Abort
	)

	s.wg = new(sync.WaitGroup)
	s.stopNowWG = new(sync.WaitGroup)

	go func() {

		s.wg.Add(1) // count this gorouting as running too

		sig := <-s.c // wait for os.Signal

		s.wg.Done() // the function is done (well not really but ...)

		// If signal received is SIGABRT, run StopNow handlers and one timeout.
		if sig == syscall.SIGABRT {
			s.waitOrTimeout(sig, s.stopNowWG, s.stopNowTimeout)
		} else {
			// Any other signal will run with two timeouts.
			s.waitOrTimeout(sig, s.wg, s.stopTimeout)
		}

	}()

}

// Registers a signalable worker.
func (s *Server) AddWorker(worker Worker) {
	s.workers = append(s.workers, worker)
}

// Deregisters a signalable worker.
func (s *Server) RemoveWorker(worker Worker) {
	// TODO: remove worker from list
	// not used now
}

// Notify a single worker (or service, luckily soju.Service implements soju.Worker)
func (s *Server) notify(sig os.Signal, worker Worker) {

	// Create a new notifier.
	d := &DefaultDoneNotifier{}

	// If signal is SIGABRT, then add the notifier to the stopNow waiting group.
	if sig == syscall.SIGABRT {
		d.wg = s.stopNowWG
		s.stopNowNotifiers = append(s.stopNowNotifiers, d)
	} else {
		// Any other signal goes to the common waiting group.
		d.wg = s.wg
		s.doneNotifiers = append(s.doneNotifiers, d)
	}

	// Add 1 to the waiting group.
	d.wg.Add(1)

	// Worker methods must be called in a goroutine.
	// If not, the shutdowns are serialized and if one of them hang the whole server hangs.
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
		// Graceful stop
		go worker.Stop(d)
	case syscall.SIGABRT:
		// Abort now! Timeout has passed!
		go worker.StopNow(d)
	}

	return

}

// Notify all workers.
func (s *Server) NotifyWorkers(sig os.Signal) {
	for i := range s.workers {
		s.notify(sig, s.workers[i])
	}
	return
}

// Serve initializes the server (setting the timeouts) and starts listening for
// OS signals. It returns an exit code when the service is stopped.
func (s *Server) Serve(stopTimeout, stopNowTimeout time.Duration) int {

	s.stopTimeout = stopTimeout
	s.stopNowTimeout = stopNowTimeout

	// Initialize the server only once.
	s.initialized.Do(s.initialize)

	// waits on waitinggroup (forever)
	s.wg.Wait()

	// waits on return code channel
	return <-s.end

}
