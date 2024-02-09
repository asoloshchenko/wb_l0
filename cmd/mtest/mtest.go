package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type test struct {
	OrderUID    string `json:"order_uid" db:"order_uid"`
	TrackNumber string `json:"track_number" db:"track_number"`

	Name  string `json:"name" db:"name"`
	Phone string `json:"phone" db:"phone"`
}

func main() {

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", "postgres", "12345", "localhost", "5432", "wb")
	db, _ := pgxpool.New(context.Background(), conString)
	err := db.Ping(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()

	rows, err := db.Query(ctx, "SELECT order_uid, track_number, name, phone FROM public.orders WHERE order_uid = $1", "cuxCLBoCJRpAHRqIuiJE")

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := pgx.CollectRows(rows, pgx.RowToStructByName[test])

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(msgs)
	db.Close()
}
