package publisher

import (
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

type Msg struct {
	OrderUID    string `json:"order_uid" fake:"{regex:[a-zA-Z]{20}}"`
	TrackNumber string `json:"track_number" fake:"{regex:[a-zA-Z]{15}}"`
	Entry       string `json:"entry" fake:"{randomstring:[WBIL, TEST, UNNAMED]}"`
	Delivery    struct {
		Name    string `json:"name" fake:"{name}"`
		Phone   string `json:"phone" fake:"{phone}"`
		Zip     string `json:"zip" fake:"{zip}"`
		City    string `json:"city" fake:"{city}"`
		Address string `json:"address" fake:"{street}"`
		Region  string `json:"region" fake:"{state}"`
		Email   string `json:"email" fake:"{email}"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction" fake:"-"`
		RequestID    string `json:"request_id"`
		Currency     string `json:"currency" fake:"{currencyshort}"`
		Provider     string `json:"provider" fake:"{randomstring:[wbpay, greenpay, yellowpay]}"`
		Amount       int    `json:"amount" fake:"{intrange:100,99999}"`
		PaymentDt    int    `json:"payment_dt" fake:"{intrange:1637900000,1638907727}"`
		Bank         string `json:"bank" fake:"{randomstring:[wb, green, yellow, blue]}"`
		DeliveryCost int    `json:"delivery_cost" fake:"{intrange:100,1000}"`
		GoodsTotal   int    `json:"goods_total" fake:"{intrange:100,500}"`
		CustomFee    int    `json:"custom_fee" fake:"{intrange:0,1000}"`
	} `json:"payment"`
	Items             []Item    `json:"items" fakesize:"1,6"`
	Locale            string    `json:"locale" fake:"{randomstring:[en,ru,de,es,fr]}"`
	InternalSignature string    `json:"internal_signature" fake:"-"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id" fake:"{intrange:10,9999}"`
	DateCreated       time.Time `json:"date_created" fake:"{date}"`
	OofShard          string    `json:"oof_shard" fake:"{regex:[0-9]{1,2}}"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" fake:"{intrange:1000000,9999999}"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price" fake:"{intrange:30,99999}"`
	Rid         string `json:"rid"`
	Name        string `json:"name" fake:"{productname}"`
	Sale        int    `json:"sale" fake:"{intrange:30,9999}"`
	Size        string `json:"size" fake:"{randomstring:[0,S,M,L,XL,XXL]}"`
	TotalPrice  int    `json:"total_price " fake:"{intrange:30,99999}"`
	NmID        int    `json:"nm_id" fake:"{intrange:1000000,9999999}"`
	Brand       string `json:"brand" fake:"{company}"`
	Status      int    `json:"status" fake:"{intrange:200,202}"`
}

func GetMsg() Msg {
	var f Msg
	gofakeit.Struct(&f)
	f.Payment.Transaction = f.OrderUID
	f.Delivery.Phone = "+" + f.Delivery.Phone

	for i := range f.Items {
		f.Items[i].TrackNumber = f.TrackNumber
	}

	fmt.Printf("Fake struct: %+v\n", f)

	//m, _ := json.Marshal(f)

	//fmt.Printf("Published message: %s\n", m)

	return f

}
