package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"

	gcppubsub "cloud.google.com/go/pubsub"
)

type Publisher struct {
	client *gcppubsub.Client
}

type JobEventPayload struct {
	UserID      string `json:"userId"`
	JobID       string `json:"jobId"`
	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"companyName"`
	TraceID     string `json:"traceId"`
}

func NewPublisher(ctx context.Context, projectID string) (*Publisher, error) {
	client, err := gcppubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Publisher{client: client}, nil
}

func (p *Publisher) Publish(topicID string, payload any) {
	go func() {
		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("pubsub marshal", "topic", topicID, "error", err)
			return
		}
		ctx := context.Background()
		result := p.client.Topic(topicID).Publish(ctx, &gcppubsub.Message{Data: data})
		if _, err := result.Get(ctx); err != nil {
			slog.Error("pubsub publish", "topic", topicID, "error", err)
		}
	}()
}

func (p *Publisher) Close() {
	p.client.Close()
}
