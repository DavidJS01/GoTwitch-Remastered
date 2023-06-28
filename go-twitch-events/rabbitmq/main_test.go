package rabbitmq

import (
	"fmt"
	"os"
	s "strings"
	"testing"
)

func TestConnectToRabbitMQ(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skipping integration test")
	}
	rabbitMQURL := fmt.Sprintf("amqp://guest:guest@%s:%s/", os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT"))
	rabbitMQAddress := s.Split(rabbitMQURL, "@")[1]
	t.Setenv("AMQP_SERVER_URL", rabbitMQURL)
	t.Log(os.Getenv("AMQP_SERVER_URL"))

	conn := ConnectToRabbitMQ()
	if conn.RemoteAddr().String()+"/" != rabbitMQAddress {
		t.Errorf("Connection address does not match rabbitmq url, expected %s got %s", rabbitMQAddress, conn.RemoteAddr().String())
	}

}
