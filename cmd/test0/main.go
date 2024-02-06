package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/stan.go"
)

func main() {
	//TODO: init config ?
	//TODO: init logger
	//TODO: init connection to db
	//TODO: init cache
	//TODO: subscribe
	//TODO: run service
	//TODO: Graceful shutdown
	fmt.Println("start")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	nc, _ := stan.Connect("test-cluster", "1")
	nc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	<-done
	fmt.Println("stop")

}
