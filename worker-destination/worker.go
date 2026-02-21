package main

import (
	"encoding/json"
	"log"

	//	"github.com/baelthebard42/Hulaak/worker-destination/config"
	worker_nats "github.com/baelthebard42/Hulaak/worker-destination/nats"
	"github.com/baelthebard42/Hulaak/worker-destination/utils"
	// "github.com/baelthebard42/Hulaak/worker-destination/utils"
)

func main() {

	//	cfg := config.Load()

	log.Println("Worker-destination initated...")

	postgres, err := utils.NewDBConnection("postgres://anjal:anjal@localhost:5432/postgres?sslmode=disable")

	if err != nil {
		log.Println("error connecting to database %v", err)
		return
	}

	NATS, err := worker_nats.NewNATSConnection("nats://localhost:4222")

	if err != nil {
		log.Println("error connecting to NATS client %v", err)
		return
	}

	sub, err := NATS.SubscribeWithDurableConsumer("webhook_events", "worker-destination", "DELIVERIES")

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

			var event utils.Payload

			err := json.Unmarshal(msg.Data, &event)

			if err != nil {
				log.Println("error marshalling received event: %v", err)
				continue
			}

			webhook, endpoint_url, err := utils.FetchDeliveryDetails(postgres, event.Delivery_id)

			if err != nil {
				log.Println("error fetching delivery details: %v", err)
				continue
			}

			err = utils.SendWebhook(*webhook, endpoint_url)

			if err != nil {
				log.Println("error sending webhook: %v", err)

				err = utils.UpdateAfterError(postgres, event.Delivery_id, err.Error())
				if err != nil {
					log.Println("error updating delivery status: %v", err)
				}
				continue
			}

			err = utils.UpdateAfterSuccess(postgres, event.Delivery_id)

			if err != nil {
				log.Println("error updating delivery status: %v", err)
				continue
			}
			msg.Ack()

		}

	}

}
