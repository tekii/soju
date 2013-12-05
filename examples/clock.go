package examples

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tekii/soju"
)

type ClockService struct {
	Called  bool
	Started bool
	Logger  *log.Logger
	LogFile *os.File
}

func (this *ClockService) Start() (err error) {

	//Initialize a log file for the service
	this.LogFile, err = os.OpenFile("tmp/clockwork.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	this.Logger = log.New(this.LogFile, "ClockService > ", log.Ldate|log.Ltime|log.Lmicroseconds)
	this.Logger.Printf("Starting ClockService")

	//Register a handle
	http.HandleFunc("/clock", func(w http.ResponseWriter, r *http.Request) {

		now := time.Now()

		this.Logger.Printf("Service called, the time is: %s", now.String())

		w.Write([]byte(now.String()))

		return

	})

	//Start listening for service calls
	go func() {
		err = http.ListenAndServe("localhost:9111", nil)
		if err != nil {
			return
		}
	}()

	this.Started = true

	return

}

func (this *ClockService) Stop() (err error) {

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
