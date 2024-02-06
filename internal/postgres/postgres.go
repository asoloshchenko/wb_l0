package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"test/internal/cache"

	_ "github.com/jackc/pgx/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(dbName, dbAddr, dbPort, dbUsername, dbPassword string) (*Storage, error) {

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUsername, dbPassword, dbAddr, dbPort, dbName)

	db, err := sql.Open("pgx", conString)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) WriteMessage(msg cache.DataStruct) error {
	tx, err := s.db.BeginTx(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(`INSERT INTO orders (order_uid, track_number, entry,
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
		tx.Rollback()
		return err
	}

	for _, it := range msg.Items {
		_, err = tx.Exec(`INSERT INTO items (order_uid, chrt_id, track_number,
								 price, rid, name, sale, size, total_price,
								 nm_id, brand, status)
						 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
								 $9, $10, $11, $12)`,
			msg.OrderUID, it.ChrtID, it.TrackNumber,
			it.Price, it.Rid, it.Name, it.Sale, it.Size,
			it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
