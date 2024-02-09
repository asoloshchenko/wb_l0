package postgres

import (
	"context"
	"log/slog"

	//"database/sql"
	"fmt"

	"time"

	"github.com/asoloshchenko/wb_l0/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(dbName, dbAddr, dbPort, dbUsername, dbPassword string) (*Storage, error) {
	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUsername, dbPassword, dbAddr, dbPort, dbName)

	db, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// WriteMessage writes order to the postgres database.
//
// It takes a context and a Order pointer as parameters and returns an error.
func (s *Storage) WriteMessage(ctx context.Context, msg *model.Order) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO public.orders
							    (order_uid, track_number, entry,
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
		tx.Rollback(ctx)
		return err
	}

	for _, it := range msg.Items {
		_, err = tx.Exec(ctx, `INSERT INTO public.items (order_uid, chrt_id, track_number,
								 price, rid, name, sale, size, total_price,
								 nm_id, brand, status)
						 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
								 $9, $10, $11, $12)`,
			msg.OrderUID, it.ChrtID, it.TrackNumber,
			it.Price, it.Rid, it.Name, it.Sale, it.Size,
			it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		tx.Rollback(context.TODO())
		return err
	}

	return nil
}

func (s *Storage) GetMessageByID(id string) (model.Order, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var msg model.Order

	err := s.db.QueryRow(ctx, `SELECT order_uid, track_number, entry,
									  name, phone, zip, city, address,
									  region, email, transaction, request_id,
									  currency, provider, amount, payment_dt,
									  bank, delivery_cost, goods_total, 
									  custom_fee, locale, internal_signature,
									  customer_id, delivery_service, shardkey,
	                                  sm_id, date_created, oof_shard
							     FROM public.orders 
								WHERE order_uid = $1`, id).Scan(
		&msg.OrderUID,
		&msg.TrackNumber,
		&msg.Entry,
		&msg.Delivery.Name,
		&msg.Delivery.Phone,
		&msg.Delivery.Zip,
		&msg.Delivery.City,
		&msg.Delivery.Address,
		&msg.Delivery.Region,
		&msg.Delivery.Email,
		&msg.Payment.Transaction,
		&msg.Payment.RequestID,
		&msg.Payment.Currency,
		&msg.Payment.Provider,
		&msg.Payment.Amount,
		&msg.Payment.PaymentDt,
		&msg.Payment.Bank,
		&msg.Payment.DeliveryCost,
		&msg.Payment.GoodsTotal,
		&msg.Payment.CustomFee,
		&msg.Locale,
		&msg.InternalSignature,
		&msg.CustomerID,
		&msg.DeliveryService,
		&msg.Shardkey,
		&msg.SmID,
		&msg.DateCreated,
		&msg.OofShard)

	if err != nil {
		//slog.Error(err.Error())
		return model.Order{}, err
	}

	rows, err := s.db.Query(ctx, `SELECT chrt_id, track_number,
										 price, rid, name, sale, 
										 size, total_price,
										 nm_id, brand, status
								    FROM public.items 
								   WHERE track_number = $1`, msg.TrackNumber)
	if err != nil {
		//slog.Error(err.Error())
		return model.Order{}, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Item])

	if err != nil {
		slog.Error(err.Error())
		return model.Order{}, err
	}

	msg.Items = append(msg.Items, items...)

	return msg, nil
}

func (s *Storage) GetCachedMessages() (map[string]model.Order, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, _ := s.db.Query(ctx, `SELECT order_uid, track_number, entry,
	name, phone, zip, city, address,
	region, email, transaction, request_id,
	currency, provider, amount, payment_dt,
	bank, delivery_cost, goods_total, 
	custom_fee, locale, internal_signature,
	customer_id, delivery_service, shardkey,
	sm_id, date_created, oof_shard FROM public.orders`)
	orders := make(map[string]model.Order)
	var msg model.Order

	pgx.ForEachRow(rows, []any{&msg.OrderUID,
		&msg.TrackNumber,
		&msg.Entry,
		&msg.Delivery.Name,
		&msg.Delivery.Phone,
		&msg.Delivery.Zip,
		&msg.Delivery.City,
		&msg.Delivery.Address,
		&msg.Delivery.Region,
		&msg.Delivery.Email,
		&msg.Payment.Transaction,
		&msg.Payment.RequestID,
		&msg.Payment.Currency,
		&msg.Payment.Provider,
		&msg.Payment.Amount,
		&msg.Payment.PaymentDt,
		&msg.Payment.Bank,
		&msg.Payment.DeliveryCost,
		&msg.Payment.GoodsTotal,
		&msg.Payment.CustomFee,
		&msg.Locale,
		&msg.InternalSignature,
		&msg.CustomerID,
		&msg.DeliveryService,
		&msg.Shardkey,
		&msg.SmID,
		&msg.DateCreated,
		&msg.OofShard},
		func() error {
			orders[msg.TrackNumber] = msg
			return nil
		})
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Order])

	if err != nil {
		return nil, err
	}
	tmp := make(map[string]model.Order)

	for _, record := range records {
		tmp[record.TrackNumber] = record
	}

	rows, _ = s.db.Query(ctx, `SELECT chrt_id, track_number,
									  price, rid, name, sale, 
									  size, total_price,
									  nm_id, brand, status 
								FROM public.items`)

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Item])
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		entry := tmp[it.TrackNumber]
		entry.Items = append(entry.Items, it)
		tmp[it.TrackNumber] = entry
	}

	// reassigning
	for _, v := range tmp {
		tmp[v.OrderUID] = v
		delete(tmp, v.TrackNumber)
	}

	return tmp, nil
}
