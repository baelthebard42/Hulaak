package worker

import (
	"context"
	"encoding/json"

	"log"
	"time"

	control_nats "github.com/baelthebard42/Hulaak/control/internal/nats"
	"github.com/baelthebard42/Hulaak/control/internal/outbox"
)

type Runner struct {
	repo *outbox.Repository
	nats *control_nats.NATS
}

func NewRunner(repo *outbox.Repository, nats_conn *control_nats.NATS) *Runner {
	return &Runner{repo: repo, nats: nats_conn}
}

func (r *Runner) processJob(ctx context.Context, job control_nats.WebhookDeliveryEvent) {

	job_json, err := json.Marshal(job)

	//fmt.Println("processing this job: %v", string(job_json))

	if err != nil {
		log.Println("Error converting webhook data to json: %v", err)
	}

	err = r.nats.PublishEvent("webhook.delivery", job_json)

	if err != nil {
		log.Fatalln("Error sending outbox event to NATS: %v", err)
		r.repo.MarkAsUnprocessed(ctx, job.Delivery_id) //setting status back to unprocessed if delivery failed
	}

	err = r.repo.DeleteThisDelivery(ctx, job.Delivery_id)

	if err != nil {
		log.Println("Error deleting published delivery: %v", err)
	}

	return

}

func (r *Runner) processBatch(ctx context.Context, batch_size int) {

	jobs, err := r.repo.ClaimOutboxBatch(ctx, batch_size)

	if err != nil {
		log.Println(err)
		return
	}

	for _, job := range jobs {
		r.processJob(ctx, job)

	}

	return

}

func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second) // ticker is a channel that receives a value every second

	defer ticker.Stop()

	log.Println("Starting background worker for publishing events from outbox table...")

	for {
		select { // waits for whichever event happens first
		case <-ctx.Done(): // shutdown requested i.e cancel() (returned in the main.go) is called
			return

		case <-ticker.C:
			r.processBatch(ctx, 50)
		}
	}
}
