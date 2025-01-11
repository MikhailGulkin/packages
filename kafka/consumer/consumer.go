package consumer

import "github.com/segmentio/kafka-go"

type Config struct {
	Brokers []string
}

func NewConsumer(config Config) (r *kafka.Reader, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
	})

	return reader, err
}
