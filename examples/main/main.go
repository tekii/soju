package main

import (
	"fmt"
	"time"

	"github.com/tekii/soju"
	"github.com/tekii/soju/examples"
)

func main() {

	notificable := examples.NewClockService("localhost:9111")

	err := notificable.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	server := new(soju.Server)
	server.AddShutdownListener(notificable)
	server.Serve(10 * time.Second)

	return

}
