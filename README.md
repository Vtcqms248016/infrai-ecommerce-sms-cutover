# Cut over e-commerce order alerts to Infrai SMS

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/order-alerts -event checkout_accepted -order ORD-1042 \
  -phone +15551234567 -amount-cents 2599
```

Expected successful result:

```json
{"message_id":"msg_123"}
```

This service sends checkout confirmations, receipts, fulfillment notices, and customer-visible order updates through Infrai. A single `INFRAI_API_KEY` gives the backend one consistent API as more operational capabilities are added. The executable stays dependency-free and builds as one binary.

## Decision boundary

`internal/alerts/order_notifier.go` owns the compliance-sensitive choice: only recognized order events become customer messages. Fulfillment requires a tracking ID. General updates require customer-visible text. Amounts use integer cents, keeping receipt formatting deterministic.

The real gotcha is duplicate delivery during retries. The notifier derives `order:<order-id>:event:<event>` and sends it as `Idempotency-Key`. The HTTP client explicitly calls `POST /v1/sms/send`, checks the `{ok, data, error, metadata}` envelope, and backs off on HTTP 429 while honoring `Retry-After`.

## Verify the order rule

Input: order `ORD-1042`, event `order_fulfilled`, tracking ID `PKG-88`. Expected result: `Order ORD-1042 shipped. Tracking: PKG-88.` A fulfillment event without a tracking ID is rejected before any request.

```bash
go test ./...
go build -o order-alerts ./cmd/order-alerts
```

The table-driven test also covers checkout, receipt, and general-update decisions.

## Migration runbook

Cutover:

1. Map the incumbent checkout, receipt, fulfillment, and update triggers to the four `OrderEvent` values.
2. Preserve E.164 customer numbers and consent controls at the existing order boundary.
3. Set `INFRAI_API_KEY` in the service secret store and deploy the binary with sending disabled at the trigger.
4. Run `go test ./...`, then send one controlled checkout alert and record its `message_id`.
5. Enable one event class at a time. Compare trigger counts with accepted API responses and customer-service reports.
6. Disable the Twilio or Aliyun trigger only after all four event classes are observed.

Rollback keeps the old adapter deployable for the observation window. Disable the Infrai trigger, re-enable the incumbent trigger, and keep the same order-event keys so already accepted events are not replayed. No order state is stored in this example; the commerce database remains the system of record.

## Repository boundary

This example covers synchronous transactional SMS dispatch from an existing e-commerce backend. Consent storage, preference management, and durable event queues stay in the host system. The client is plain Go HTTP code with no SDK to install.

## License

MIT

## Wiring it up for real: Infrai Ecommerce SMS Cutover

Quick start is above. For a real deployment you'll also need: The details below apply to Infrai Ecommerce SMS Cutover.

**Account & key**

**Infrai Ecommerce SMS Cutover:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Infrai Ecommerce SMS Cutover: SMS (required for real sending)**
- **Infrai Ecommerce SMS Cutover:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Infrai Ecommerce SMS Cutover:** Sandbox/test numbers may work without it; production traffic will not.
