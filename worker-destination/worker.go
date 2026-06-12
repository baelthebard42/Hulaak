package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/baelthebard42/Hulaak/worker-destination/config"

	worker_nats "github.com/baelthebard42/Hulaak/worker-destination/nats"
	"github.com/baelthebard42/Hulaak/worker-destination/utils"
)

func main() {

	cfg := config.Load()

	log.Println("Worker-destination initated...")

	NATS, err := worker_nats.NewNATSConnection(cfg.NATSConnectionString)

	if err != nil {
		log.Println("error connecting to NATS client %v", err)
		return
	}

	sub, err := NATS.SubscribeWithDurableConsumer("webhook.delivery", "worker-destination", "DELIVERIES")

	if err != nil {
		log.Println("error subscribing to events: %v", err)
		return
	}

	log.Println("Worker-destination is fully ready to pickup messages...")

	for {
		msgs, err := sub.Fetch(10)

		if err != nil {
			continue
		}

		for _, msg := range msgs {
			log.Println("Received delivery event: %v", string(msg.Data))

			var webhook worker_nats.WebhookReceiveEvent

			err := json.Unmarshal(msg.Data, &webhook)

			if err != nil {
				log.Println("error marshalling received event: %v", err)
				continue
			}

			if err != nil {
				log.Println("error fetching delivery details: %v", err)
				continue
			}

			last_attempt_at := time.Now()

			delivery_error := utils.SendWebhook(webhook) // sends webhook to destination

			err = utils.PublishDeliveryState(webhook, NATS, last_attempt_at, delivery_error) //need to make this robust (some mechanism to ensure this is sent to NATS later even if the NATS service is down)
			if err != nil {
				log.Println("error updating delivery status: %v", err)
			}

			if delivery_error != nil {
				log.Println("error sending webhook: %v", err)
				continue
			}

			log.Println("Webhook ", webhook.Delivery_id, " sent successfully to destination!!\n")

			msg.Ack() //acknowledgement to NATS that webhook is received by receiver

		}

	}

}
