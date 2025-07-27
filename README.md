# CHAT

A real-time group chat application with room-based messaging and persistence support.

The central server contains a hub and client components, where the client is a representation of a websocket connection. The hub acts as a coordinator for all client events including room entry/exit, message broadcasting, and state persistence.

## Setup

1. Set environment and start the database with docker

```bash
# copy env
cp .env.example .env

# start database container
docker compose up db -d
```

The database instance starts on the port specified in the env file

2. Run the server:

```bash
go run cmd/main.go
```

The server starts on `localhost:8000` by default

## Usage

To join a room, connect to `/ws/{userId}` with query parameters:

-   `roomId` - ID of the room to join
    ie: `ws://localhost:8000/ws/{userId}?roomId={room1}`

## TODO

-   [x] Add room support
-   [x] Persist rooms and messages in database
-   [ ] Add user authentication
