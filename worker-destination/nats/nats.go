package worker_nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type NATS struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewNATSConnection(connection_string string) (*NATS, error) {

	nc, err := nats.Connect(
		connection_string,
		nats.MaxReconnects(-1),            // retry indefinitely
		nats.ReconnectWait(2*time.Second), // wait between retries
	)

	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()

	if err != nil {
		return nil, err
	}

	return &NATS{conn: nc, js: js}, nil

}

func (n *NATS) Close() {
	n.conn.Close()
}

func (n *NATS) PublishEvent(subject string, payload json.RawMessage) error {

	_, err := n.js.Publish(subject, payload)

	fmt.Println("Sent event:", string(payload))

	return err

}
