package postgres

import (
	"context"
	//"database/sql"
	"fmt"

	"test/internal/model"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(dbName, dbAddr, dbPort, dbUsername, dbPassword string) (*Storage, error) {
	// TODO: connection pool

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUsername, dbPassword, dbAddr, dbPort, dbName)

	db, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	err = db.Ping(context.TODO())
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) WriteMessage(msg model.DataStruct) error {
	tx, err := s.db.BeginTx(context.TODO(), pgx.TxOptions{})

	if err != nil {
		return err
	}

	_, err = tx.Exec(context.TODO(), `INSERT INTO orders (order_uid, track_number, entry,
								 name, phone, zip, city, address,
								 region, email, transaction, request_id,
								 currency, provider, amount, payment_dt,
								 bank, delivery_cost, goods_total, 
								 custom_fee, locale, internal_signature,
								 customer_id, delivery_service, shardkey,
								 sm_id, date_created, oof_shard)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
						     $9, $10, $11, $12, $13, $14, $15,
							 $16, $17, $18, $19, $20, $21, $22,
							 $23, $24, $25, $26, $27, $28)`,
		msg.OrderUID, msg.TrackNumber, msg.Entry,
		msg.Delivery.Name, msg.Delivery.Phone, msg.Delivery.Zip,
		msg.Delivery.City, msg.Delivery.Address, msg.Delivery.Region,
		msg.Delivery.Email, msg.Payment.Transaction, msg.Payment.RequestID,
		msg.Payment.Currency, msg.Payment.Provider, msg.Payment.Amount,
		msg.Payment.PaymentDt, msg.Payment.Bank, msg.Payment.DeliveryCost,
		msg.Payment.GoodsTotal, msg.Payment.CustomFee, msg.Locale,
		msg.InternalSignature, msg.CustomerID, msg.DeliveryService,
		msg.Shardkey, msg.SmID, msg.DateCreated, msg.OofShard)
	if err != nil {
		tx.Rollback(context.TODO())
		return err
	}

	for _, it := range msg.Items {
		_, err = tx.Exec(context.TODO(), `INSERT INTO items (order_uid, chrt_id, track_number,
								 price, rid, name, sale, size, total_price,
								 nm_id, brand, status)
						 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
								 $9, $10, $11, $12)`,
			msg.OrderUID, it.ChrtID, it.TrackNumber,
			it.Price, it.Rid, it.Name, it.Sale, it.Size,
			it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			tx.Rollback(context.TODO())
			return err
		}
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		tx.Rollback(context.TODO())
		return err
	}

	return nil
}

func (s *Storage) GetMessageByID(id string) (model.DataStruct, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, _ := s.db.Query(ctx, "SELECT * FROM orders WHERE order_uid = $1", id)
	msgs, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.DataStruct])

	if err != nil {
		return model.DataStruct{}, err
	}

	rows, _ = s.db.Query(ctx, "SELECT * FROM items WHERE track_number = $1", msgs[0].TrackNumber)

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Item])

	if err != nil {
		return model.DataStruct{}, err
	}

	msgs[0].Items = append(msgs[0].Items, items...)

	return msgs[0], nil
}

func (s *Storage) GetCachedMessages() (map[string]model.DataStruct, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, _ := s.db.Query(ctx, "SELECT * FROM orders")
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.DataStruct])

	if err != nil {
		return nil, err
	}
	tmp := make(map[string]model.DataStruct)

	for _, record := range records {
		tmp[record.TrackNumber] = record
	}

	rows, _ = s.db.Query(ctx, "SELECT * FROM items")

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Item])
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		entry := tmp[it.TrackNumber]
		entry.Items = append(entry.Items, it)
		tmp[it.TrackNumber] = entry
	}

	if err != nil {
		return nil, err
	}
	return tmp, nil
}
