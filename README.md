
# Beacon : Realtime Engine

## About

Beacon is a real-time messaging and broadcasting service built with **Go**, **WebSockets**, **Redis**, and **Docker**.

I built Beacon to explore concurrent backend engineering and event-driven architecture, including WebSocket connection management, selective message fanout, Redis-based event delivery, and durable message history.

The service uses **Redis Pub/Sub** for ephemeral real-time delivery and **Redis Streams** for durable message history, while a concurrent Go WebSocket hub manages connected clients and channel subscriptions.

## 🧠 Core Concepts & Key Takeaways

### 1. Decoupled Event-Driven Architecture

Beacon is designed as an independent real-time service rather than being tightly coupled to a primary web application or database.

- **Redis Pub/Sub:** Acts as an asynchronous event bridge between message producers and the WebSocket engine.
- **Independent Components:** Publishers do not need to know which WebSocket clients are connected, while Beacon independently consumes events and delivers them to relevant clients.
- **Failure Isolation:** Live Pub/Sub delivery and durable Stream persistence are intentionally decoupled so a history write failure does not automatically prevent real-time message delivery.

A simplified message flow looks like:

```text
Producer
   │
   ▼
Redis Pub/Sub
   │
   ▼
Beacon
   │
   ▼
WebSocket Hub
   │
   ▼
Subscribed Clients
```

### 2. Concurrent WebSocket Hub

Each WebSocket connection uses separate **ReadPump** and **WritePump** goroutines.

```text
Client
├── ReadPump  → receives commands/messages
└── WritePump → sends messages to the WebSocket
```

This allows WebSocket reads and writes to progress independently without one connection blocking unrelated clients.

A central **Hub goroutine** owns shared connection and subscription state. Client goroutines communicate with the Hub through Go channels instead of modifying shared maps directly.

The Hub processes events such as:

- Client registration
- Client disconnection
- Channel subscription
- Channel unsubscription
- Message broadcasting

This follows an event-loop style concurrency model: many goroutines can produce events concurrently, while a single Hub goroutine serialises mutations to shared state.

### 3. Channel Subscriptions & Selective Fanout

Clients can dynamically subscribe and unsubscribe from logical messaging channels over their WebSocket connection.

For example:

```json
{"type":"subscribe","channel":"sports"}
```

```json
{"type":"unsubscribe","channel":"sports"}
```

The Hub maintains bidirectional subscription indexes:

```text
Channels
channel → connected clients

ClientChannels
client → subscribed channels
```

This allows Beacon to route a message only to clients subscribed to its channel rather than scanning and broadcasting to every connected client.

```text
sports message
      │
      ▼
Hub.Channels["sports"]
      │
   ┌──┴──┐
   ▼     ▼
Client A Client C
```

Client disconnections also trigger automatic cleanup of all associated channel memberships.

### 4. Redis Pub/Sub for Live Delivery

Redis Pub/Sub handles ephemeral real-time event delivery.

Beacon subscribes to Redis and forwards incoming messages through the WebSocket Hub to clients subscribed to the corresponding application channel.

Pub/Sub is intentionally used for **live delivery**, not durable storage. If a client is disconnected when an event is published, it cannot recover that event from Pub/Sub alone.

### 5. Redis Streams for Durable History

Beacon also persists messages using **Redis Streams** and `XADD`.

Each stored event receives a Redis Stream ID, which Beacon exposes as `Message.ID`.

This provides durable history alongside the ephemeral Pub/Sub path:

```text
Message
   ├──► Redis Pub/Sub → live WebSocket delivery
   │
   └──► Redis Stream  → durable history
```

The `/history` API allows clients to retrieve previously stored messages after reconnecting.

### 6. History Filtering & Cursor Pagination (In Progress)

The history API supports retrieving messages by channel and navigating message history using Redis Stream IDs as cursors.

Instead of repeatedly requesting only the latest fixed set of messages, clients can paginate through history using `before` and `after` cursors.

Supported capabilities:

- Filter history by channel
- Limit the number of returned messages
- Use `Message.ID` (Redis Stream ID) as a pagination cursor
- Load older messages using `before`
- Load newer messages using `after`

Examples:

```text
GET /history?channel=sports&limit=20

GET /history?channel=sports&limit=20&before=<stream-id>

GET /history?channel=sports&limit=20&after=<stream-id>
```

This allows clients to incrementally load message history without repeatedly fetching the same latest messages.

### 7. Container Optimization

Beacon is containerized using a **multi-stage Docker build**.

- **Builder Stage:** Uses the Go toolchain to compile the application into a standalone binary.
- **Runtime Stage:** Copies only the compiled binary and required runtime files into a lightweight Alpine Linux image.
- **Result:** The final production image remains under **20 MB**, reducing deployment size and unnecessary runtime tooling.

## 📂 Project Structure

```text
realtime-engine/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── websocket/
│   │   ├── client.go
│   │   ├── hub.go
│   │   └── subscription.go
│   └── ...
├── .dockerignore
├── Dockerfile
├── go.mod
└── go.sum
```

## 🛠️ How to Run and Test Locally

Beacon is configured using environment variables so runtime configuration remains separate from application code.

### 1. Start Redis

If Redis is installed locally through Homebrew:

```bash
brew services start redis
```

### 2. Build the Docker Image

```bash
docker build -t realtime-engine .
```

### 3. Run Beacon

```bash
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e REDIS_ADDR=host.docker.internal:6379 \
  -e REDIS_CHANNEL=global_activity \
  realtime-engine
```

### 4. Connect WebSocket Clients

Open multiple terminal windows to simulate independent clients:

```bash
npx wscat --connect ws://localhost:8080/ws
```

Subscribe one client to `sports`:

```json
{"type":"subscribe","channel":"sports"}
```

Subscribe another client to `general`:

```json
{"type":"subscribe","channel":"general"}
```

### 5. Send a Message

A WebSocket client can publish a message using:

```json
{
  "type": "message",
  "channel": "sports",
  "data": {
    "text": "Hello sports!"
  }
}
```

Only clients subscribed to `sports` should receive the resulting live message.

### 6. Test Unsubscription

```json
{"type":"unsubscribe","channel":"sports"}
```

Future `sports` messages should no longer be delivered to that connection.

### 7. Retrieve Message History

Retrieve recent history using the `/history` endpoint:

```text
GET /history?channel=sports&limit=20
```

Use the returned Redis Stream IDs to paginate through older or newer history:

```text
GET /history?channel=sports&limit=20&before=<stream-id>

GET /history?channel=sports&limit=20&after=<stream-id>
```

## 🏗️ Architecture Summary

```text
                         ┌─────────────────┐
                         │ Message Producer│
                         └────────┬────────┘
                                  │
                         ┌────────▼────────┐
                         │      Redis      │
                         │ Pub/Sub + Stream│
                         └────────┬────────┘
                                  │
                                  ▼
                         ┌─────────────────┐
                         │     Beacon      │
                         │                 │
                         │ Redis Listener  │
                         │       │         │
                         │       ▼         │
                         │      Hub        │
                         └───────┬─────────┘
                                 │
                     selective fanout
                        ┌────────┼────────┐
                        ▼        ▼        ▼
                     Client A Client B Client C
```

Beacon separates **live event delivery**, **durable history**, **WebSocket connection management**, and **subscription routing** while using Go's goroutines and channels to coordinate concurrent connections.

