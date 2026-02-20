package main

import (
	"encoding/json"
	"log"

	"github.com/baelthebard42/Hulaak/worker-nats/config"
	worker_nats "github.com/baelthebard42/Hulaak/worker-nats/nats"
	"github.com/baelthebard42/Hulaak/worker-nats/utils"
)

func main() {

	cfg := config.Load()

	log.Println("Worker-NATS initated...")

	postgres, err := utils.NewDBConnection(cfg.DatabaseURL)

	if err != nil {
		log.Fatalln("error connecting to database %v", err)
		return
	}

	NATS, err := worker_nats.NewNATSConnection(cfg.NATSConnectionString)

	if err != nil {
		log.Fatalln("error connecting to NATS client %v", err)
		return
	}

	log.Println("Worker beginning to claim and send from outbox..")

	for {

		//log.Println("Claiming batch...")

		delivery_batch, err := utils.ClaimOutboxBatch(postgres)

		if err != nil {
			log.Fatalln("error fetching delivery data: %v", err)
		}

		//print("Retrieved data: ")

		//	log.Println("Sending batch...")

		for _, value := range delivery_batch {
			//	fmt.Printf("type of %v: %T", index, value)

			delivery_json, err := json.Marshal(value)

			if err != nil {
				log.Fatalln("error converting delivery to json: %v", err)

			}

			err = NATS.PublishEvent("webhook_event", delivery_json)

			if err != nil {
				log.Fatalln("Error sending outbox event to NATS: %v", err)
				utils.MarkAsNull(postgres, value.Delivery_id) //setting status back to null if delivery failed
			}

			err = utils.MarkAsPublished(postgres, value.Delivery_id)

			if err != nil {
				panic("Error marking sent messages as sent in database!!!!!")
			}

		}

	}

}
