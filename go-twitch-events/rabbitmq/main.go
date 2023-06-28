package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
)

func ConnectToRabbitMQ() *amqp.Connection {
	// Get rabbitMQ server url from env
	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	// Create a new RabbitMQ connection
	rabbitMQConnection, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	return rabbitMQConnection

}

func ConnectToRabbitMQChannel(connectRabbitMQ *amqp.Connection) *amqp.Channel {
	channelConnection, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	return channelConnection
}

func CreateMessage(body []byte) amqp.Publishing {
	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}
	return message

}
