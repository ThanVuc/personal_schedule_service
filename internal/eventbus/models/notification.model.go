package event_models

type Notification struct {
	ID              *string  `json:"id,omitempty" bson:"_id,omitempty"`
	Title           string   `json:"title" bson:"title"`
	Message         string   `json:"message" bson:"message"`
	Link            *string  `json:"link,omitempty" bson:"link,omitempty"`
	SenderID        string   `json:"sender_id" bson:"sender_id"`
	ReceiverIDs     []string `json:"receiver_ids" bson:"receiver_ids"`
	ImageURL        *string  `json:"img_url,omitempty" bson:"img_url,omitempty"`
	CorrelationID   string   `json:"correlation_id" bson:"correlation_id"`
	CorrelationType int      `json:"correlation_type" bson:"correlation_type"`
}
