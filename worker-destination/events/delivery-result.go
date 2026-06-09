package events

type DeliveryResultEvent struct {
	Delivery_id string `json:"delivery_id"`
	Succeeded   bool   `json:"succeeded"`
}
