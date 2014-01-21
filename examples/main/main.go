package main

import (
	"fmt"
	"time"

	"github.com/tekii/soju"
	"github.com/tekii/soju/examples"
)

func main() {

	signalable := examples.NewClockService("localhost:9111")

	server := new(soju.Server)
	server.SetService(signalable)
	err := signalable.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	server.Serve(10*time.Second, 10*time.Second)

	return

}
