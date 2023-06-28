package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	z "strings"
	"testing"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		fmt.Print(message)
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestParseUserName(t *testing.T) {
	mock_message := ":username!username@username.tmi.twitch.tv PRIVMSG #katevolved :message here"
	username := parseUserName(mock_message)
	if username != "username" {
		t.Errorf("Parsing a username was incorrect, expected 'username' got '%s'", username)
	}
}

func TestParseMessage(t *testing.T) {
	mock_message := ":username!username@username.tmi.twitch.tv PRIVMSG #katevolved :message here"
	message := parseMessage(mock_message)
	if message != "message here" {
		t.Errorf("Parsing a message was incorrect, expected 'message here' got '%s'", message)
	}
}

func TestCreateWebsocketClient(t *testing.T) {
	// creates bytes buffer for panic test
	s := httptest.NewServer(http.HandlerFunc(echo))
	address := z.Split(s.URL, "http://")[1]
	defer s.Close()

	// test for web socket error
	_, err := createWebSocketClient("bad_address", "ws")
	if err == nil {
		t.Error("Expected error when creating ws client with bad address")
	}

	// test for web socket without error
	client, _ := createWebSocketClient(address, "ws")
	if client.RemoteAddr().String() != address {
		t.Errorf("Error connecting test ws client to 127.0.0, server url %s, client url %s", address, client.RemoteAddr().String())
	}
	s.Close()
}

func TestAuthenticateClient(t *testing.T) {
	// temporarily modify environment variables
	t.Setenv("twitchAuth", "password")
	t.Setenv("twitchUsername", "username")

	// create websocket server and client
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()
	address := z.Split(s.URL, "http://")[1]
	client, _ := createWebSocketClient(address, "ws")
	authenticateClient(client, "twitch_channel_here")

	// test for `PASS <oauth_password> message`
	_, msg, _ := client.ReadMessage()
	fmt.Print(string(msg))
	if string(msg) != fmt.Sprintf("PASS %s", os.Getenv("twitchAuth")) {
		t.Errorf("expected to recieve `PASS %s`, got %s", os.Getenv("twitchAuth"), string(msg))
	}

	// test for `NICK <username> message`
	_, msg, _ = client.ReadMessage()
	fmt.Print(string(msg))
	if string(msg) != fmt.Sprintf("NICK %s", os.Getenv("twitchUsername")) {
		t.Errorf("expected to recieve NICK %s, got %s", os.Getenv("twitchUsername"), string(msg))
	}

	// test for `JOIN #` message
	_, msg, _ = client.ReadMessage()
	fmt.Print(string(msg))
	if string(msg) != "JOIN #twitch_channel_here" {
		t.Errorf("expected to recieve JOIN #twitch_channel_here, got %s", string(msg))
	}
}

func mockInsertTwitchMessage(username string, message string, channel string) {}

func TestParseTwitchMessage(t *testing.T) {
	// create messages to test on
	mockTwitchMessage := []byte("username!username@username.tmi.twitch.tv PRIVMSG #katevolved :message here")
	mockPingMessage := []byte("PING server message here")

	// start websocket server for testing
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// create client websocket conn
	address := z.Split(s.URL, "http://")[1]
	client, _ := createWebSocketClient(address, "ws")

	// run twitch message test
	parsedUsername, parsedMessage := parseTwitchMessage(mockTwitchMessage, "Katevolved", client, mockInsertTwitchMessage) // TODO: finish unit test

	if parsedUsername != "username" {
		fmt.Print(len(parsedUsername))
		t.Errorf("Unexpected parsed twitch message, expected username 'username' got %s", parsedUsername)
	}

	if parsedMessage != "message here" {
		fmt.Print(len(parsedUsername))
		t.Errorf("Unexpected parsed twitch message, expected message 'message here' got %s", parsedMessage)
	}

	// run ping message test
	parsedUsername, parsedMessage = parseTwitchMessage(mockPingMessage, "Katevolved", client, mockInsertTwitchMessage)
	if parsedUsername != "" && parsedMessage != "" {
		t.Errorf("Expected null parsed username when recieving PING message, got %s, %s", parsedUsername, parsedMessage)
	}

	_, msg, _ := client.ReadMessage()
	if string(msg) != "PONG :tmi.twitch.tv" {
		t.Errorf("Unexpected PING response, expected PONG got %s", string(msg))
	}

}
