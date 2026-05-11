package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"

	gcppubsub "cloud.google.com/go/pubsub"
)

type Publisher struct {
	client *gcppubsub.Client
	topic  *gcppubsub.Topic
}

type JobCreatedPayload struct {
	UserID      string `json:"userId"`
	JobID       string `json:"jobId"`
	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"companyName"`
	TraceID     string `json:"traceId"`
}

func NewPublisher(ctx context.Context, projectID, topicID string) (*Publisher, error) {
	client, err := gcppubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		client: client,
		topic:  client.Topic(topicID),
	}, nil
}

func (p *Publisher) PublishJobCreated(payload JobCreatedPayload) {
	go func() {
		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("pubsub marshal job created", "error", err)
			return
		}
		ctx := context.Background()
		result := p.topic.Publish(ctx, &gcppubsub.Message{Data: data})
		if _, err := result.Get(ctx); err != nil {
			slog.Error("pubsub publish job created", "error", err)
		}
	}()
}

func (p *Publisher) Close() {
	p.client.Close()
}
