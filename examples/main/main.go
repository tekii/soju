package main

import (
	"fmt"
	"time"

	"github.com/tekii/soju"
	"github.com/tekii/soju/examples"
)

func main() {

	notificable := new(examples.ClockService)

	err := notificable.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	server := new(soju.Server)
	server.AddShutdownListener(notificable)
	server.Serve(5 * time.Second)

	return

}
