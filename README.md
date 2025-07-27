# CHAT

A real-time group chat application with room-based messaging support.

The central server contains a hub and client components, where the client is a representation of a websocket connection. The hub acts as a central coordinator for all client events such as `register`, `unregister` and `broadcast`.

## Setup

Run the server:

```bash
go run cmd/main.go
```

The server starts on `localhost:8000` by default

## Usage

To join a room, connect to `/ws/{userId}` with query parameters:

-   `roomId` - ID of the room to join
-   `username` - Display name of the user
    ie: `ws://localhost:8000/ws/{userId}?roomId={room1}&username={mrshabel}`

## TODO

-   [x] Add room support
-   [ ] Persist rooms and messages in database
-   [ ] Add user authentication
