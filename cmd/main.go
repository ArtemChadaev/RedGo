package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	go func() {
		fmt.Println("starting server")
	}()

	<-ch

	fmt.Println("stopping server")
}
