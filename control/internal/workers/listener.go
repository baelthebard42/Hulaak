package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/baelthebard42/Hulaak/control/internal/events"
	control_nats "github.com/baelthebard42/Hulaak/control/internal/nats"
)

type Listener struct {
	repo *events.Repository
	nats *control_nats.NATS
}

func NewListener(repo *events.Repository, nats_conn *control_nats.NATS) *Listener {
	return &Listener{repo: repo, nats: nats_conn}
}

func (l *Listener) HandleDeliveryEvent(ctx context.Context, event control_nats.DeliveryResultEvent) error {

	deliveryState, err := l.repo.GetDeliveryStateFromDID(ctx, event.Delivery_id)

	if err != nil {
		log.Println("Error fetching delivery details: ", err)
		return err
	}

	if deliveryState.Status == "success" || deliveryState.Status == "failed" || !event.Last_attempt.After(deliveryState.Last_attempt_at) {

		log.Println("Duplicate event detected...dropping the event.", deliveryState)
		return fmt.Errorf("Duplicate event detected.")
	}

	log.Println("Received event for update: (id, success): ", event.Delivery_id, event.Succeeded)

	err = l.repo.UpdateAfterEvent(ctx, event)

	if err != nil {
		log.Println("Error updating delivery table..", err)
		return err
	}

	return nil
}

func (l *Listener) Run(ctx context.Context) {
	var ErrDuplicateEvent = errors.New("duplicate event detected")
	ticker := time.NewTicker(time.Second)

	defer ticker.Stop()

	sub, err := l.nats.SubscribeWithDurableConsumer("webhook.delivery.state", "control", "DELIVERIES")

	if err != nil {
		log.Println("Error subscribing for events..", err)
		return
	}

	log.Println("Starting background listener for receiving events..")

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:

			msgs, err := sub.Fetch(10)

			if err != nil {
				continue
			}
			for _, msg := range msgs {
				log.Println("Received eventkkkk: ", string(msg.Data))
				var event control_nats.DeliveryResultEvent
				err := json.Unmarshal(msg.Data, &event)

				if err != nil {
					log.Println("error marshalling received event: %v", err)
					continue
				}

				err = l.HandleDeliveryEvent(ctx, event)
				if err != nil && !errors.Is(err, ErrDuplicateEvent) {
					log.Println("error handling received event:", err)
					continue // no ack → redeliver & retry
				}
				msg.Ack() // ack on success OR duplicate

			}

		}
	}
}
