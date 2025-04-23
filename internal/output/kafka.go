package output

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaOutput struct {
	writer *kafka.Writer
	topic  string
	ctx    context.Context
}

func NewKafkaOutput(ctx context.Context, brokers []string, topic string, clientID string) *KafkaOutput {
	return &KafkaOutput{
		topic: topic,
		ctx:   ctx,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (o *KafkaOutput) Send(event map[string]interface{}) error {
	select {
	case <-o.ctx.Done():
		return o.ctx.Err()
	default:
		msg := kafka.Message{
			Value: []byte(toJSONString(event)),
		}
		return o.writer.WriteMessages(context.Background(), msg)
	}
}
