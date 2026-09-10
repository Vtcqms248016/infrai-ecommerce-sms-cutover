package alerts

import "testing"

func TestMessageForOrderDecision(t *testing.T) {
	tests := []struct {
		name    string
		update  OrderUpdate
		want    string
		wantErr bool
	}{
		{"checkout confirmation", OrderUpdate{OrderID: "ORD-1042", Phone: "+15551234567", Event: CheckoutAccepted, AmountCents: 2599}, "Order ORD-1042 is confirmed. Total: USD 25.99.", false},
		{"receipt", OrderUpdate{OrderID: "ORD-1042", Phone: "+15551234567", Event: ReceiptIssued, AmountCents: 2599}, "Receipt for order ORD-1042 is ready. Paid: USD 25.99.", false},
		{"fulfillment", OrderUpdate{OrderID: "ORD-1042", Phone: "+15551234567", Event: OrderFulfilled, TrackingID: "PKG-88"}, "Order ORD-1042 shipped. Tracking: PKG-88.", false},
		{"customer update", OrderUpdate{OrderID: "ORD-1042", Phone: "+15551234567", Event: OrderUpdated, Summary: "delivery moved to Friday"}, "Order ORD-1042 update: delivery moved to Friday", false},
		{"fulfillment requires tracking", OrderUpdate{OrderID: "ORD-1042", Phone: "+15551234567", Event: OrderFulfilled}, "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MessageFor(test.update)
			if (err != nil) != test.wantErr {
				t.Fatalf("MessageFor() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("MessageFor() = %q, want %q", got, test.want)
			}
		})
	}
}
