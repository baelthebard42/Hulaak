package control_nats

import (
	"encoding/json"

	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NATS struct {
	conn *nats.Conn
	// js   nats.JetStreamContext
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

	//js, err := nc.JetStream()

	if err != nil {
		return nil, err
	}

	log.Println("control successfully connected to NATS client!!")

	log.Printf("NATS conn: %v", nc != nil)
	//log.Printf("JetStream enabled check: %+v", js)

	return &NATS{conn: nc}, nil

}

func (n *NATS) Close() {
	n.conn.Close()
	return
}

func (n *NATS) js() (nats.JetStreamContext, error) {
	return n.conn.JetStream()
}

func (n *NATS) PublishEvent(subject string, payload json.RawMessage) error {
	js, err := n.js()
	if err != nil {
		return err
	}

	_, err = js.Publish(subject, payload)
	return err
}

func (n *NATS) EnsureStream(streamName string) error {
	js, err := n.js()
	_, err = js.StreamInfo(streamName)
	if err == nil {
		return nil
	}

	if err == nats.ErrStreamNotFound {
		log.Println("Stream not %v found, creating it...", streamName)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"webhook.>"},
			Storage:  nats.FileStorage,
		})
		return err
	}

	return err
}

func (n *NATS) EnsureConsumer(streamName string, consumerName string) error {
	js, err := n.js()
	_, err = js.ConsumerInfo(streamName, consumerName)
	if err == nil {
		return nil
	}

	if err == nats.ErrConsumerNotFound {
		log.Println("Consumer not %v found, adding..", streamName)
		_, err = js.AddConsumer(streamName, &nats.ConsumerConfig{
			Durable:    consumerName,
			AckPolicy:  nats.AckExplicitPolicy,
			MaxDeliver: 10,
			BackOff: []time.Duration{
				1 * time.Minute,
				2 * time.Minute,
				4 * time.Minute,
				8 * time.Minute,
				16 * time.Minute,
				30 * time.Minute,
				60 * time.Minute,
			},
		})
		return err
	}

	return err
}

func (n *NATS) SubscribeWithDurableConsumer(subject string, durableConsumerName string, streamName string) (*nats.Subscription, error) {

	js, err := n.js()
	err = n.EnsureStream(streamName)

	if err != nil {
		log.Fatalln("Error ensuring stream: %v", err)
		return nil, err
	}

	err = n.EnsureConsumer(streamName, durableConsumerName)

	if err != nil {
		log.Fatalln("Error ensuring consumer: %v", err)
		return nil, err
	}

	sub, err := js.PullSubscribe(
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
