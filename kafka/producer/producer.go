package producer

import "github.com/segmentio/kafka-go"

type Config struct {
	Brokers []string
	Topic   string
}

func NewProducer(config Config) (*kafka.Writer, error) {
	w := &kafka.Writer{
		Addr:     kafka.TCP(config.Brokers...),
		Topic:    config.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return w, nil
}
