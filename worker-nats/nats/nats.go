package worker_nats

import (
	"encoding/json"
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

	log.Println("Worker-NATS successfully connected to NATS client!!")

	return &NATS{conn: nc, js: js}, nil

}

func (n *NATS) Close() {
	n.conn.Close()
}

func (n *NATS) PublishEvent(subject string, payload json.RawMessage) error {

	ack, err := n.js.Publish(subject, payload)

	log.Println("Sent event:", string(payload))
	log.Println("Acknowledgement:", ack.Sequence)

	return err

}
