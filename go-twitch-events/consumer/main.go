package main

import (
	"fmt"
	"log"
	"os/exec"
	s "strings"
	mq "test.com/m/go-twitch-events/rabbitmq"
	"test.com/m/internal/database"
)

func listenStream(stream string) {
	log.Printf("Connecting to the chat for stream %s", stream)
	// https://serverfault.com/a/903631
	// start listening to the stream
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo $$; exec ./stream %s", stream))
	// insert stream event status with listening bool, pid, and stream name
	database.InsertStreamEventStatus(true, cmd.Process.Pid, stream)
	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}
}

func main() {
	// connect to rabbitmq
	rabbitMQConnection := mq.ConnectToRabbitMQ()
	defer rabbitMQConnection.Close()
	// connect to rabbitmq channel
	rabbitMQChannel := mq.ConnectToRabbitMQChannel(rabbitMQConnection)
	_, err := rabbitMQChannel.QueueDeclare(
		"Streams", // queue name
		true,      // durable
		false,     // auto delete
		false,     // exclusive
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		panic(err)
	}
	defer rabbitMQChannel.Close()

	// Subscribe to Streams for consuming messages.
	messages, err := rabbitMQChannel.Consume(
		"Streams", // queue name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		log.Println(err)
	}

	log.Println("Successfully connected to RabbitMQ")
	log.Println("Waiting for messages")

	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)

	go func() {
		for message := range messages {
			log.Printf(" > Received message: %s\n", message.Body)
			// convert event's message.Body into string, either "start" or "stop"
			command := string(message.Body)

			if s.Contains(command, "start") {
				// if the string contains the start command, call listenStream
				stream := s.Split(command, " ")[1]
				listenStream(stream)
			}
			if s.Contains(command, "stop") {
				// parse stream name from message
				stream := s.Split(command, " ")[1]

				// get the latest PID used to listen to that stream
				pid := database.GetLatestPID(stream)

				// kill the process associated with that PID
				cmd := exec.Command("kill", fmt.Sprint(pid))
				err := cmd.Run()
				if err != nil {
					log.Print(err)
				}

				// update the stream events status table and mark listening as false
				database.UpdateStreamEventStatus(pid, stream)
				if err != nil {
					log.Print(err)
				}
			}

		}
	}()

	<-forever
}
