package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test/internal/cache"
	"test/internal/model"

	"test/internal/config"
	"test/internal/postgres"

	"github.com/nats-io/stan.go"
)

func main() {
	//TODO: init config ?
	slog.Info("Reading config...")
	cfg := config.ReadConfig()

	//slog.Info()
	//TODO: init connection to db
	slog.Info("Coonection to db...")
	storage, err := postgres.New(cfg.DbName, cfg.DbAddr, cfg.DbPort, cfg.DbUsername, cfg.DbPassword)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	//TODO: init cache
	c := cache.New(time.Minute*10, time.Minute*5) //TODO: read time from config
	err = c.Restore(storage)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	//TODO: subscribe
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	nc, err := stan.Connect("test-cluster", "1") //TODO: read name from config
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("start")
	nc.Subscribe("foo", func(m *stan.Msg) {
		var msg model.DataStruct

		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		if msg.OrderUID == "" {
			slog.Info("OrderUID is empty")
			return
		}

		c.Set(msg.OrderUID, msg, 0)
		storage.WriteMessage(msg)
		//fmt.Println(len(msg.Items))

		//fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	<-done
	fmt.Println("stop")

}
