package main

import (
	"fmt"
	"github.com/google/uuid"

	//"fmt"
	"github.com/MikhailGulkin/packages/rabbit"
)

func main() {
	conn, err := rabbit.NewRabbitCh(rabbit.Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "user.messages",
		QueuePattern: "user.id",
	})
	if err != nil {
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	err = conn.DeclareAndBindQueue(uuid.New().String(), "")
	if err != nil {
		return
	}
	consume, err := conn.Consume(
		"",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return
	}

	for {
		msg, ok := <-consume
		if !ok {
			return
		}
		fmt.Println(string(msg.Body))
		//err := msg.Ack(false)
		//if err != nil {
		//	fmt.Println(err)
		//}
	}
}
