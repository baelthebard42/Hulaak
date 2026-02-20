package main

import (
	"log"

	"github.com/baelthebard42/Hulaak/worker-destination/config"
	worker_nats "github.com/baelthebard42/Hulaak/worker-destination/nats"
	// "github.com/baelthebard42/Hulaak/worker-destination/utils"
)

func main() {

	cfg := config.Load()

	log.Println("Worker-destination initated...")

	//postgres, err := utils.NewDBConnection(cfg.DatabaseURL)
	var err error

	if err != nil {
		log.Fatalln("error connecting to database %v", err)
		return
	}

	NATS, err := worker_nats.NewNATSConnection(cfg.NATSConnectionString)

	if err != nil {
		log.Fatalln("error connecting to NATS client %v", err)
		return
	}

	sub, err := NATS.SubscribeWithDurableConsumer("webhook_events", "worker-destination", "DELIVERIES")

	if err != nil {
		log.Fatalln("error subscribing to events: %v", err)
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

			//if sending to destination succeeds
			msg.Ack()
		}

	}

}
