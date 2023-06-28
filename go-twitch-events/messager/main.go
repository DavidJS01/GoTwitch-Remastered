package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	mq "test.com/m/go-twitch-events/rabbitmq"
)

type MessageResponse struct {
	State       string `json:"state"`
	Channel     string `json:"channel"`
	Status_Code int    `json:"status_code"`
}

func WriteMessage(body []byte) error {
	// connect to rabbitmq
	rabbitMQConnection := mq.ConnectToRabbitMQ()
	defer rabbitMQConnection.Close()
	
	// connect to rabbitmq channel
	rabbitMQChannel := mq.ConnectToRabbitMQChannel(rabbitMQConnection)
	defer rabbitMQChannel.Close()

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

	message := mq.CreateMessage(body)

	// Attempt to publish a message to the queue.
	if err := rabbitMQChannel.Publish(
		"",        // exchange
		"Streams", // queue name
		false,     // mandatory
		false,     // immediate
		message,   // message to publish
	); err != nil {
		return err
	}

	return nil
}

func messageResponse(state string, channel string, status_code int) MessageResponse {
	var response MessageResponse
	response.State = state
	response.Channel = channel
	response.Status_Code = status_code

	return response
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	state := mux.Vars(r)["state"]
	channel := mux.Vars(r)["channel"]
	response := messageResponse(state, channel, 200)
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Print(err)
	}
	WriteMessage([]byte(fmt.Sprintf("%s %s", state, channel)))

}

func main() {

	mux := mux.NewRouter()
	mux.HandleFunc("/message", sendMessageHandler).Queries("state", "{state}", "channel", "{channel}").Methods("POST")
	err := http.ListenAndServe(":9090", mux)
	if err != nil {
		log.Print(err)
	}
}
