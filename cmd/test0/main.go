package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test/internal/config"
	"test/internal/postgres"

	"github.com/nats-io/stan.go"
)

func main() {
	//TODO: init config ?
	cfg := config.ReadConfig()

	slog.Info()
	//TODO: init connection to db

	storage, err := postgres.New(cfg.DbName, cfg.DbAddr, cfg.DbPort, cfg.DbUsername, cfg.DbPassword)
	//TODO: init cache
	c := cache.New(24*time.Hour, 0)
	//TODO: subscribe
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	nc, err := stan.Connect("test-cluster", "1")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("start")
	nc.Subscribe("foo", func(m *stan.Msg) {
		var msg cache.DataStruct

		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(len(msg.Items))

		//fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	<-done
	fmt.Println("stop")

}
