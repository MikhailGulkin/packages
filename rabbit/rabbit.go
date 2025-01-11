package rabbit

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Conn struct {
	*amqp.Connection
	*amqp.Channel

	cfg Config
}
type Config struct {
	URL          string
	Exchange     string
	QueuePattern string
}

func NewRabbitCh(config Config) (*Conn, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Conn{
		Connection: conn,
		Channel:    ch,
		cfg:        config,
	}, nil
}

func (c *Conn) DeclareAndBindQueue(uniqueKey, eName string) error {
	if eName == "" {
		eName = c.cfg.Exchange
	}
	_, err := c.QueueDeclare(
		fmt.Sprintf("%s.%s", c.cfg.QueuePattern, uniqueKey),
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = c.QueueBind(
		fmt.Sprintf("%s.%s", c.cfg.QueuePattern, uniqueKey),
		fmt.Sprintf("%s.%s", c.cfg.QueuePattern, uniqueKey),
		eName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) Close() error {
	if err := c.Channel.Close(); err != nil {
		return err
	}
	if err := c.Connection.Close(); err != nil {
		return err
	}
	return nil
}
