The `go-twitch-events` folder contains all code that interacts with RabbitMQ.

### Folder Layout
```
├── go-twitch-events
│   ├── consumer
│   │   └── main.go
│   ├── messager
│   │   └── main.go
│   ├── rabbitmq
│   │   ├── main.go
│   │   └── main_test.go
│   └── README.md
```

#### Consumer
The `consumer.go` file listens for two requests from the broker:
1. `start {stream}`
    - attempt to open connection to a livestream given a stream name
2. `stop {stream}`
    - close a connection to a livestream given a stream name

#### Messager
The `messager.go` file sends requests to the broker. It opens a REST server on port 9090, and the user of the service may send two requests:

1. `start {stream}`
2. `stop {stream}`
