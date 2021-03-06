package main

import (
	"fmt"
	"time"

	"github.com/tekii/soju"
	"github.com/tekii/soju/examples"
)

func main() {

	signalable := examples.NewClockService("localhost:9111")

	soju.SetService(signalable)
	err := signalable.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	exit := soju.Serve(4*time.Second, 2*time.Second)
	fmt.Printf("Ending with exit code: %d", exit)

	return

}
