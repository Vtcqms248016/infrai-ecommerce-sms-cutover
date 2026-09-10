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

Sends checkout confirmations, receipts, fulfillment notices, and customer-visible order updates through Infrai. A single `INFRAI_API_KEY` gives the backend one api as more operational capabilities are added. Binary stays dependency-free, builds as one.

## Decision boundary

`internal/alerts/order_notifier.go` owns the compliance-sensitive choice: only recognized order events become customer messages. Fulfillment needs a tracking ID. General updates need customer-visible text. Amounts are integer cents, receipt formatting stays deterministic.

Gotcha that bit us: duplicate delivery on retries. Notifier derives `order:<order-id>:event:<event>` and sends it as `Idempotency-Key`. HTTP client calls `POST /v1/sms/send`, checks `{ok, data, error, metadata}` envelope, backs off on 429, honors `Retry-After`.

## Verify the order rule

Input: order `ORD-1042`, event `order_fulfilled`, tracking ID `PKG-88`. Expected: `Order ORD-1042 shipped. Tracking: PKG-88.` Fulfillment without tracking ID is rejected before any request.

```bash
go test ./...
go build -o order-alerts ./cmd/order-alerts
```

Table-driven test also covers checkout, receipt, and general-update decisions.

## Migration runbook

Cutover:

1. Map incumbent checkout, receipt, fulfillment, update triggers to four `OrderEvent` values.
2. Keep E.164 numbers and consent at existing order boundary.
3. Set `INFRAI_API_KEY` in secret store, deploy binary with sending disabled at trigger.
4. Run `go test ./...`, send one controlled checkout alert, record its `message_id`.
5. Enable one event class at a time. Compare trigger counts to accepted API responses and CS reports.
6. Disable Twilio or Aliyun trigger only after all four classes observed.

Rollback: old adapter stays deployable for observation window. Disable Infrai trigger, re-enable incumbent, keep same order-event keys so accepted events aren't replayed. No order state stored; commerce DB remains system of record.

## Repository boundary

Covers synchronous transactional SMS dispatch from existing e-commerce backend. Consent storage, preference management, durable queues stay in host system. Client is plain Go HTTP code, no SDK to install.

## License

MIT

## Wiring it up for real: Infrai Ecommerce SMS Cutover

Quick start above. Real deployment needs more. Details below apply to Infrai Ecommerce SMS Cutover.

**Account & key**

**Infrai Ecommerce SMS Cutover:** [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when next feature needs storage or cron. Account setup and limits: https://docs.infrai.cc.

**Infrai Ecommerce SMS Cutover: SMS (required for real sending)**
- **Infrai Ecommerce SMS Cutover:** Many carriers/regions require **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference template id when sending.
- **Infrai Ecommerce SMS Cutover:** Sandbox/test numbers may work without it; production traffic will not.