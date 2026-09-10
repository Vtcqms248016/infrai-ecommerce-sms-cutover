package alerts

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/infrai-ecommerce-sms-cutover/internal/infrai"
)

type OrderEvent string

const (
	CheckoutAccepted OrderEvent = "checkout_accepted"
	ReceiptIssued    OrderEvent = "receipt_issued"
	OrderFulfilled   OrderEvent = "order_fulfilled"
	OrderUpdated     OrderEvent = "order_updated"
)

type OrderUpdate struct {
	OrderID     string
	Phone       string
	Event       OrderEvent
	AmountCents int64
	TrackingID  string
	Summary     string
}

type Sender interface {
	Send(context.Context, infrai.SendRequest, string) (infrai.SendResult, error)
}

type Notifier struct {
	sender Sender
}

func NewNotifier(sender Sender) *Notifier {
	return &Notifier{sender: sender}
}

func (n *Notifier) Notify(ctx context.Context, update OrderUpdate) (infrai.SendResult, error) {
	message, err := MessageFor(update)
	if err != nil {
		return infrai.SendResult{}, err
	}
	key := fmt.Sprintf("order:%s:event:%s", update.OrderID, update.Event)
	return n.sender.Send(ctx, infrai.SendRequest{To: update.Phone, Message: message}, key)
}

func MessageFor(update OrderUpdate) (string, error) {
	if strings.TrimSpace(update.OrderID) == "" || strings.TrimSpace(update.Phone) == "" {
		return "", fmt.Errorf("order ID and phone are required")
	}

	switch update.Event {
	case CheckoutAccepted:
		return fmt.Sprintf("Order %s is confirmed. Total: %s.", update.OrderID, formatAmount(update.AmountCents)), nil
	case ReceiptIssued:
		return fmt.Sprintf("Receipt for order %s is ready. Paid: %s.", update.OrderID, formatAmount(update.AmountCents)), nil
	case OrderFulfilled:
		if strings.TrimSpace(update.TrackingID) == "" {
			return "", fmt.Errorf("tracking ID is required for fulfillment")
		}
		return fmt.Sprintf("Order %s shipped. Tracking: %s.", update.OrderID, update.TrackingID), nil
	case OrderUpdated:
		if strings.TrimSpace(update.Summary) == "" {
			return "", fmt.Errorf("summary is required for an order update")
		}
		return fmt.Sprintf("Order %s update: %s", update.OrderID, update.Summary), nil
	default:
		return "", fmt.Errorf("unsupported order event %q", update.Event)
	}
}

func formatAmount(cents int64) string {
	return fmt.Sprintf("USD %d.%02d", cents/100, cents%100)
}
