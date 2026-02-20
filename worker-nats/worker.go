package main

import (
	"encoding/json"
	"log"

	"github.com/baelthebard42/Hulaak/worker-destination/config"
	worker_nats "github.com/baelthebard42/Hulaak/worker-destination/nats"
	"github.com/baelthebard42/Hulaak/worker-destination/utils"
)

func main() {

	cfg := config.Load()

	postgres, err := utils.NewDBConnection(cfg.DatabaseURL)

	NATS, err := worker_nats.NewNATSConnection(cfg.NATSConnectionString)

	if err != nil {
		log.Fatalln("error connecting to NATS client %v", err)
		return
	}

	if err != nil {
		log.Fatalln("error connecting to database %v", err)
		return
	}

	for {

		log.Println("Claiming batch...")

		delivery_batch, err := utils.ClaimOutboxBatch(postgres)

		if err != nil {
			log.Fatalln("error fetching delivery data: %v", err)
		}

		//print("Retrieved data: ")

		log.Println("Sending batch...")

		for _, value := range delivery_batch {
			//	fmt.Printf("type of %v: %T", index, value)

			delivery_json, err := json.Marshal(value)

			if err != nil {
				log.Fatalln("error converting delivery to json: %v", err)
			}

			err = NATS.PublishEvent("webhook_event", delivery_json)

			if err != nil {
				log.Fatalln("Error sending outbox event to NATS: %v", err)
			}

			err = utils.MarkAsPublished(postgres, value.Delivery_id)

			if err != nil {
				panic("Error marking sent messages as sent in database!!!!!")
			}

		}

	}

}
