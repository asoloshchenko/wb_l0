package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	p "test/internal/publisher"
	"time"

	stan "github.com/nats-io/stan.go"
	//"log/slog"
	"fmt"
)

func main() {

	// TODO add router

	sc, err := stan.Connect("test-cluster", "pub")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer sc.Close()

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		msg := p.GetMsg()
		bytes, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err.Error())
			io.WriteString(w, err.Error())
		}

		err = sc.Publish("foo", bytes)
		if err != nil {
			fmt.Println(err.Error())
			io.WriteString(w, err.Error())
		}

		fmt.Println("sent succsesfully")
		io.WriteString(w, "OK")

	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("Failed to start server:", err.Error())
		}
	}()

	<-done
	fmt.Println("stopping server")

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//TODO: Graceful shutdown
}
