In development.

Notes so I can make beautiful documentation later:

- Goose installation instructions
    - Goose env vars config
- Golangci-lint 
    - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## .env
Create a .env file with the below values:
```
twitchUsername= YOUR TWITCH USERNAME HERE
twitchClientId= YOUR CLIENT ID HERE
twitchClientSecret= YOUR CLIENT SECRET HERE
twitchAuth= YOUR OAUTH TOKEN HERE
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="user=postgres dbname=postgres sslmode=disable host=172.17.0.1 port=5432 password=postgres"


```