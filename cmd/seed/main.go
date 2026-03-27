package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MonsieurTib/service-bus-tui/internal/azure"
)

const emulatorConnectionString = "Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true"

var sampleMessages = []azure.MessageInfo{
	{Subject: "order.created", ContentType: "application/json", Body: `{"orderId":"ORD-001","customer":"alice@example.com","amount":42.50,"items":["widget","gadget"]}`, Properties: map[string]any{"env": "dev", "version": "1"}},
	{Subject: "order.updated", ContentType: "application/json", Body: `{"orderId":"ORD-002","status":"processing","updatedAt":"2026-03-25T10:00:00Z"}`, Properties: map[string]any{"env": "dev", "version": "1"}},
	{Subject: "order.shipped", ContentType: "application/json", Body: `{"orderId":"ORD-003","trackingNumber":"TRK-789","carrier":"DHL"}`, Properties: map[string]any{"env": "dev", "region": "eu-west"}},
	{Subject: "order.cancelled", ContentType: "application/json", Body: `{"orderId":"ORD-004","reason":"out_of_stock","refund":true}`, Properties: map[string]any{"env": "dev"}},
	{Subject: "user.registered", ContentType: "application/json", Body: `{"userId":"USR-101","email":"bob@example.com","plan":"premium"}`, Properties: map[string]any{"env": "dev", "source": "web"}},
	{Subject: "user.updated", ContentType: "application/json", Body: `{"userId":"USR-102","changes":{"plan":"free","email":"carol@example.com"}}`, Properties: map[string]any{"env": "dev"}},
	{Subject: "payment.received", ContentType: "application/json", Body: `{"paymentId":"PAY-501","amount":99.99,"currency":"EUR","method":"card"}`, Properties: map[string]any{"env": "dev", "priority": "high"}},
	{Subject: "payment.failed", ContentType: "application/json", Body: `{"paymentId":"PAY-502","reason":"insufficient_funds","retryable":true}`, Properties: map[string]any{"env": "dev", "priority": "high"}},
	{Subject: "inventory.low", ContentType: "application/json", Body: `{"sku":"WIDGET-XL","remaining":3,"threshold":10,"warehouse":"WH-A"}`, Properties: map[string]any{"env": "dev", "alert": "true"}},
	{Subject: "notification.sent", ContentType: "application/json", Body: `{"notifId":"NOTIF-001","channel":"email","recipient":"dave@example.com","template":"welcome"}`, Properties: map[string]any{"env": "dev"}},
	{Subject: "report.generated", ContentType: "text/plain", Body: "Daily sales report for 2026-03-25: total orders=42, revenue=EUR 3820.00, returns=2.", Properties: map[string]any{"env": "dev", "type": "daily"}},
	{Subject: "healthcheck.ping", ContentType: "application/json", Body: `{"service":"order-service","status":"ok","uptime":86400}`, Properties: map[string]any{"env": "dev"}},
}

func main() {
	client, err := azure.NewServiceBusClientFromConnectionString(emulatorConnectionString)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	destinations := []string{"queue.1", "topic.1"}

	for _, dest := range destinations {
		fmt.Printf("Sending %d messages to %s...\n", len(sampleMessages), dest)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		progress := client.SendMessages(ctx, dest, sampleMessages, false)
		for p := range progress {
			if p.Err != nil {
				cancel()
				log.Fatalf("error sending to %s: %v", dest, p.Err)
			}
			if p.Done {
				fmt.Printf("  ✓ Sent %d/%d messages to %s\n", p.Sent, p.Total, dest)
			}
		}
		cancel()
	}

	fmt.Println("Done.")
}
