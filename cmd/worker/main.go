package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"os"

	"github.com/Mizbain-Fathima/remote-care-backend/internal/mq"
	"github.com/joho/godotenv"
)

type PurchaseEvent struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	VoucherID     string  `json:"voucher_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"payment_method"`
	UpiID         string  `json:"upi_id"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env not loaded, using system ENV")
	}

	rmqURL := os.Getenv("RABBIT_URL")
	if rmqURL == "" {
		log.Fatal("RABBIT_URL missing in env")
	}

	consumer, err := mq.NewConsumer(
		rmqURL,
		"voucher.events",
		"purchase.completed",
	)
	if err != nil {
		log.Fatal("failed to start consumer: ", err)
	}

	fmt.Println("Worker started. Listening for events...")

	err = consumer.Consume(context.Background(), func(body []byte) error {
		var ev PurchaseEvent
		if err := json.Unmarshal(body, &ev); err != nil {
			return err
		}

		fmt.Println("---- PURCHASE EVENT ----")
		fmt.Println("Transaction:", ev.TransactionID)
		fmt.Println("User:", ev.UserID)
		fmt.Println("Voucher:", ev.VoucherID)
		fmt.Println("Amount:", ev.Amount, ev.Currency)
		fmt.Println("Status:", ev.Status)
		fmt.Println("------------------------")

		// Optionally save to analytics DB, send email, etc.
		return nil
	})

	if err != nil {
		log.Fatal("consumer error:", err)
	}
}
