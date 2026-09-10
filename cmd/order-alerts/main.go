package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/example/infrai-ecommerce-sms-cutover/internal/alerts"
	"github.com/example/infrai-ecommerce-sms-cutover/internal/infrai"
)

func main() {
	event := flag.String("event", string(alerts.CheckoutAccepted), "checkout_accepted, receipt_issued, order_fulfilled, or order_updated")
	orderID := flag.String("order", "", "merchant order ID")
	phone := flag.String("phone", "", "customer phone in international format")
	amount := flag.Int64("amount-cents", 0, "charged amount in cents")
	tracking := flag.String("tracking", "", "carrier tracking ID")
	summary := flag.String("summary", "", "customer-visible order update")
	flag.Parse()

	client := infrai.NewSMSClient(os.Getenv("INFRAI_API_KEY"))
	notifier := alerts.NewNotifier(client)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result, err := notifier.Notify(ctx, alerts.OrderUpdate{
		OrderID: *orderID, Phone: *phone, Event: alerts.OrderEvent(*event),
		AmountCents: *amount, TrackingID: *tracking, Summary: *summary,
	})
	if err != nil {
		log.Fatal(err)
	}
	output, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
