package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"test/internal/cache"
	"test/internal/config"
	"test/internal/model"
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
	inCache := cache.New(time.Minute*10, time.Minute*5) //TODO: read time from config
	err = inCache.Restore(storage)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	//TODO: subscribe
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Subscribing...")
	nc, err := stan.Connect("test-cluster", "1") //TODO: read name from config
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

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
		inCache.Set(msg.OrderUID, msg, 0)
		storage.WriteMessage(msg)

	})

	slog.Info("Starting server...")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			w.Write([]byte("empty id"))
			return
		}
		if value, ok := inCache.Get(id); ok {
			response, err := json.Marshal(value)

			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			w.Write(response)
			return
		}

		if value, err := storage.GetMessageByID(id); err == nil {
			response, err := json.Marshal(value)

			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			w.Write(response)
			return
		}

		w.Write([]byte("No such id"))

		//w.Write([]byte("welcome"))
	})

	srv := &http.Server{
		Addr:    ":3333",
		Handler: r,
	}

	http.ListenAndServe(":3333", r)

	slog.Info("Server is on port 3333")

	<-done
	slog.Info("stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Info("failed to stop server")
	}

}
