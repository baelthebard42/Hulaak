package workers

import (
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

func (l *Listener) HandleDeliveryEvent(event control_nats.DeliveryResultEvent) error {

}
