module defaultuse

require (
	github.com/MikhailGulkin/packages v0.0.0-20250110140232-b6f74429ab8a
	github.com/google/uuid v1.6.0
)

require github.com/rabbitmq/amqp091-go v1.10.0 // indirect

go 1.23.4

replace github.com/MikhailGulkin/packages => ../../../
