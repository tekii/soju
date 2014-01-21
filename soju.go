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
		syscall.SIGABRT, //Abort
	)

	s.wg = new(sync.WaitGroup)

	go func() {

		s.wg.Add(1)  // count this gorouting as running too
		sig := <-s.c // wait for os.Signal

		//Got signal, notify and exit
		d := DefaultDoneNotifier{wg: s.wg}
		d.wg.Add(1)
		s.doneNotifiers = append(s.doneNotifiers, &d)
		//Stop the service asynchronously so that it doesn't hang the server
		go s.service.Stop(&d)
		//If any, notify all registered workers.
		s.NotifyWorkers(sig)

		s.wg.Done() // the function is done (well not really but ...)

		//This channel will be "signaled" when the WaitingGrup reaches 0
		waiter := make(chan int, 1)
		go func() {
			s.wg.Wait()
			waiter <- 0
		}()

		// waits for:
		// 1 - Waiter channel to be signaled. AKA all goroutines are done
		// 2 - First timeout
		select {
		case ec := <-waiter:
			// return code 0 (all routines are done)
			s.end <- ec
		case <-time.After(s.stopTimeout):

			//Timeout period exceded => Abort Now!
			d := DefaultDoneNotifier{wg: s.wg}
			d.wg.Add(1)
			go s.service.StopNow(&d)
			//Abort all workers
			s.NotifyWorkers(syscall.SIGABRT)

			//Waits for second timeout
			select {
			case <-time.After(s.stopNowTimeout):
				//No more wait... stop everything now!
				for i := range s.doneNotifiers {
					//releases the waiting group using bruteforce
					s.doneNotifiers[i].Done()
				}
				// return code 2 => timeout
				s.end <- 2
			}

		}

	}()
}

func (s *Server) AddWorker(worker Worker) {
	s.workers = append(s.workers, worker)
}

func (s *Server) RemoveWorker(worker Worker) {
	// TODO: remove worker from list
	// not used now
}

func (s *Server) NotifyWorkers(sig os.Signal) {
	for _, worker := range s.workers {
		//Create a new notifier.
		d := DefaultDoneNotifier{wg: s.wg}
		//Add 1 to the waiting group.
		d.wg.Add(1)
		s.doneNotifiers = append(s.doneNotifiers, &d)

		//Worker methods must be called in a goroutine.
		//If not, the shutdowns are serialized and if one of them hang the whole server hangs.
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			//Gracefull stop
			go worker.Stop(&d)
		case syscall.SIGABRT:
			//Abort now! Timeout has passed!
			go worker.StopNow(&d)
		}
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
