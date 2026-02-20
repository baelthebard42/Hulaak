package worker_nats

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NATS struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewNATSConnection(connection_string string) (*NATS, error) {

	log.Println("Connecting to NATS client...")

	var nc *nats.Conn
	var err error

	for {
		nc, err = nats.Connect(
			connection_string,
			nats.MaxReconnects(-1),
			nats.ReconnectWait(2*time.Second),
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				log.Printf("Disconnected from NATS: %v", err)
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Printf("Reconnected to NATS")
			}),
		)

		if err == nil {
			break
		}

		log.Printf("NATS not ready yet: %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()

	if err != nil {
		return nil, err
	}

	log.Println("Worker-Destination successfully connected to NATS client!!")

	return &NATS{conn: nc, js: js}, nil

}

func (n *NATS) Close() {
	n.conn.Close()
}

func (n *NATS) EnsureStream(streamName string) error {
	_, err := n.js.StreamInfo(streamName)
	if err == nil {
		return nil
	}

	if err == nats.ErrStreamNotFound {
		log.Println("Stream not %v found, creating it...", streamName)
		_, err = n.js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"webhook_events"},
			Storage:  nats.FileStorage,
		})
		return err
	}

	return err
}

func (n *NATS) EnsureConsumer(streamName string, consumerName string) error {
	_, err := n.js.ConsumerInfo(streamName, consumerName)
	if err == nil {
		return nil
	}

	if err == nats.ErrConsumerNotFound {
		log.Println("Consumer not %v found, adding..", streamName)
		_, err = n.js.AddConsumer(streamName, &nats.ConsumerConfig{
			Durable:    consumerName,
			AckPolicy:  nats.AckExplicitPolicy,
			MaxDeliver: 10,
			BackOff: []time.Duration{
				1 * time.Second,
				2 * time.Second,
				4 * time.Second,
				8 * time.Second,
				16 * time.Second,
				30 * time.Second,
				1 * time.Minute,
			},
		})
		return err
	}

	return err
}

func (n *NATS) SubscribeWithDurableConsumer(subject string, durableConsumerName string, streamName string) (*nats.Subscription, error) {

	err := n.EnsureStream(streamName)

	if err != nil {
		log.Fatalln("Error ensuring stream: %v", err)
		return nil, err
	}

	err = n.EnsureConsumer(streamName, durableConsumerName)

	if err != nil {
		log.Fatalln("Error ensuring consumer: %v", err)
		return nil, err
	}

	sub, err := n.js.PullSubscribe(
		subject,
		durableConsumerName,
		nats.BindStream(streamName),
	)

	if err != nil {

		log.Fatalln("error pulling subscription: %v", err)
		return nil, err

	}

	return sub, nil

}
