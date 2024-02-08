package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"

	"github.com/asoloshchenko/wb_l0/internal/cache"
	"github.com/asoloshchenko/wb_l0/internal/config"
	"github.com/asoloshchenko/wb_l0/internal/model"
	"github.com/asoloshchenko/wb_l0/internal/postgres"

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
	// err = inCache.Restore(storage)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	os.Exit(1)
	// }

	//TODO: subscribe

	slog.Info("Subscribing...")
	nc, err := stan.Connect("test-cluster", "1") //TODO: read name from config
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	nc.Subscribe("foo", func(m *stan.Msg) {
		var msg model.DataStruct
		slog.Info("Get msg")

		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		if msg.OrderUID == "" {
			slog.Info("OrderUID is empty")
			return
		}
		//fmt.Println(msg)

		err = storage.WriteMessage(msg)
		if err != nil {
			slog.Error(err.Error())
		}
		inCache.Set(msg.OrderUID, msg, 0)
		if stored, ok := inCache.Get(msg.OrderUID); ok {
			fmt.Println("stored:")
			fmt.Println(stored)
		}

		slog.Info("ended")

	})

	slog.Info("Starting server...")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)
	})

	r.Get("/api/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		slog.Info("Get msg:", slog.Any("id", id))
		if id == "" {
			w.Write([]byte("empty id"))
			return
		}
		if value, ok := inCache.Get(id); ok {
			response, err := json.Marshal(value)

			if err != nil {
				errResponse, _ := json.Marshal(model.Responce{Err: err.Error()})
				w.Write(errResponse)
				return
			}

			w.Write(response)
			return
		}

		value, err := storage.GetMessageByID(id)

		switch err {
		case nil:
			response, _ := json.Marshal(value)
			w.Write(response)
		case pgx.ErrNoRows: //pgx.ErrNoRows:
			errResponse, _ := json.Marshal(model.Responce{Err: "not found"})
			w.Write(errResponse)
			return
		default:
			errResponse, _ := json.Marshal(model.Responce{Err: err.Error()})
			w.Write(errResponse)
			return
		}

	})

	srv := &http.Server{
		Addr:    ":3333",
		Handler: r,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("failed to start server")
		}
	}()

	slog.Info("Server is on port 3333")

	<-done
	slog.Info("stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Info("failed to stop server")
		return
	}

	slog.Info("server stopped")

}
