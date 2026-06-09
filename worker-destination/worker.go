package main

import (
	"encoding/json"
	"log"

	"github.com/baelthebard42/Hulaak/worker-destination/config"
	"github.com/baelthebard42/Hulaak/worker-destination/events"
	worker_nats "github.com/baelthebard42/Hulaak/worker-destination/nats"
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

			var webhook events.WebhookReceiveEvent

			err := json.Unmarshal(msg.Data, &webhook)

			if err != nil {
				log.Println("error marshalling received event: %v", err)
				continue
			}

			if err != nil {
				log.Println("error fetching delivery details: %v", err)
				continue
			}

			//	err = utils.SendWebhook(webhook)

			if err != nil {
				log.Println("error sending webhook: %v", err)

				//	err = utils.UpdateAfterError(postgres, event.Delivery_id, err.Error())
				if err != nil {
					log.Println("error updating delivery status: %v", err)
				}
				continue
			}

			//	err = utils.UpdateAfterSuccess(postgres, event.Delivery_id)

			if err != nil {
				log.Println("error updating delivery status: %v", err)
				continue
			}
			log.Println("Not acknowledging  event so it is sent back again for retry")
			//msg.Ack()

		}

	}

}
