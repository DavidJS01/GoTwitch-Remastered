package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"regexp"
	s "strings"
	"test.com/m/internal/database"
)

func parseUserName(twitchMessage string) string {
	rx := regexp.MustCompile(`(\w*)!`)
	username := rx.FindString(twitchMessage)
	username = s.Split(username, "!")[0]
	return username
}

func parseMessage(twitchMessage string) string {
	rx := regexp.MustCompile(`#(.*?):(.*)`)
	message := rx.FindAllStringSubmatch(twitchMessage, -1)[0][2]
	message = s.Trim(message, "\n")
	return message
}

func createWebSocketClient(host string, scheme string) (*websocket.Conn, error) {
	// create url for websocket connection
	u := url.URL{Scheme: scheme, Host: host}
	log.WithFields(log.Fields{
		"URL": u.String(),
	}).Info("Preparing to open a websocket connection..")

	// create websocket client
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func authenticateClient(connection *websocket.Conn, twitchChannel string) {
	log.Print("Authenticating websocket client")
	oauth := fmt.Sprintf("PASS %s", os.Getenv("twitchAuth"))
	username := fmt.Sprintf("NICK %s", os.Getenv("twitchUsername"))

	// send oauth token to twitch
	log.Info("Sending Twitch the oauth token..")
	err := connection.WriteMessage(websocket.TextMessage, []byte(oauth))
	if err != nil {
		log.Fatalf("A fatal error ocurred while authenticating the oauth token %s", err.Error())
	}
	// send username to twitch
	log.Info("Sending Twitch the username..")
	err = connection.WriteMessage(websocket.TextMessage, []byte(username))
	if err != nil {
		log.Fatalf("A fatal error ocurred while authenticating the username token %s", err.Error())
	}
	// join a twitch channel's chat
	log.Info("Sending Twitch the channel's chat room to join..")
	connection.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("JOIN #%s", twitchChannel)))
	if err != nil {
		log.Fatalf("A fatal error ocurred while joining the Twitch channel's chat %s", err.Error())
	}
}

func parseTwitchMessage(message []byte, channel string, connection *websocket.Conn) (username string, parsedMessage string) {
	messageString := string(message)

	if s.Contains(messageString, "PRIVMSG") {
		message := parseMessage(messageString)
		username := parseUserName(messageString)
		log.Infof("%s: %s \n", username, message)
		return username, message
	}
	if s.Contains(messageString, "PING") {
		connection.WriteMessage(websocket.TextMessage, []byte("PONG :tmi.twitch.tv"))
	}
	return "", ""
}

func receiveHandler(connection *websocket.Conn, channel string) {
	for {
		// read a message
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Errorf("Error while recieving a twitch message: %s", err.Error())
		} else {
			// parse message for username and twitch chat message
			parsedUsername, parsedMessage := parseTwitchMessage(msg, channel, connection)
			// if the message contained a username and twitch message, insert content into postgres
			if parsedUsername != "" && parsedMessage != "" {
				database.InsertStreamer(channel)
				database.InsertTwitchMessage(parsedUsername, parsedMessage, channel)
			}

		}
	}
}

func StartStream(twitch_channel string) {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}
	// create websocket connection, connect to twitch websocket
	connection, err := createWebSocketClient("irc-ws.chat.twitch.tv:443", "wss")
	if err != nil {
		log.Fatalf("Error establishing web socket client: %s", err.Error())
	}
	// authenticate connection with twitch, join channel
	authenticateClient(connection, twitch_channel)
	// start listening to messages
	receiveHandler(connection, twitch_channel)
	defer connection.Close()
}

func main() {
	StartStream(os.Args[1])
}
