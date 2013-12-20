package examples

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/tekii/soju"
)

type ClockService struct {
	Called  bool
	Started bool
	Logger  *log.Logger
	LogFile *os.File

	//Wrap a http.Server and add a sync.WaitGroup to
	//track the http requests. Have a reference to the
	//listener in order to call wl.Close() on shutdown.
	Server    *http.Server
	wl        *soju.WaitListener
	waitGroup sync.WaitGroup
}

//Creates a new ClockService initializing the Server and adding
//the handlers.
func NewClockService(addr string) (service *ClockService) {

	//Create the service
	service = &ClockService{
		//And the http.Server
		Server: &http.Server{
			Addr:           addr,
			MaxHeaderBytes: http.DefaultMaxHeaderBytes,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
		},
	}

	//Convert a function into a type Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		now := time.Now()

		service.Logger.Println("service called, asking the time...")

		//Do nothing for a couple of seconds so that there are pending
		//requests when the process gets the kill signal.
		time.Sleep(2 * time.Second)

		service.Logger.Printf("the time is: %s", now.String())

		w.Write([]byte(now.String()))

		return

	})
	//Register the handler in the service http.Server
	service.Server.Handler = handler

	return

}

//Wraps a net.Listener in a WaitListener and starts serving.
func (this *ClockService) Start() (err error) {

	//Initialize a log file for the service
	this.LogFile, err = os.OpenFile("/tmp/clockwork.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	this.Logger = log.New(this.LogFile, "ClockService > ", log.Ldate|log.Ltime|log.Lmicroseconds)
	this.Logger.Printf("Starting ClockService")

	//Get a listener for the server address.
	l, err := net.Listen("tcp", this.Server.Addr)
	if err != nil {
		return
	}
	//Wrap the listener in a WaitListener.
	this.wl = &soju.WaitListener{
		Listener:  l,
		WaitGroup: &this.waitGroup,
	}

	this.Started = true

	//This call blocks, so make it in a goroutine.
	go this.Server.Serve(this.wl)

	return

}

func (this *ClockService) Stop() (err error) {

	this.Logger.Println("Closing the listener.")
	//Do not accept new requests.
	err = this.wl.Close()
	if err != nil {
		this.Logger.Printf("Error closing the listener [%s]", err.Error())
		return
	}

	this.Logger.Println("Waiting for all pending requests to finish...")
	//Block until all requests are done.
	this.waitGroup.Wait()

	this.Logger.Println("Stopping the service...")
	//Close the log file
	err = this.LogFile.Close()
	if err != nil {
		return
	}

	this.Started = false

	return

}

func (this *ClockService) Shutdown(doneNotifier soju.DoneNotifier) {

	this.Logger.Println("Shutdown signal!")

	if this.Started {
		this.Stop()
	}

	this.Called = true
	doneNotifier.Done()

	return

}
