# Beacon : Realtime Engine

## About

Beacon is a real-time broadcasting service. I built this project as a way to learn **Go**, **Websockets** and **Docker** by designing a decoupled **Event-Driven Architecture** using **Redis**.

## 🧠 Core Concepts & Key Takeaways

### 1. Decoupled Architecture (The Event Bridge)
This service is built as an independent microservice. It is completely separated from the main web application and the primary database (like Express and MongoDB).
* **How it communicates:** Instead of talking directly to other apps, it uses **Redis Pub/Sub** as a middleman.
* **The Flow:** The main backend acts as the *producer* (publishing events to Redis). This Go engine acts as the *consumer* (subscribing to Redis to catch those events). This keeps the apps completely independent—if the backend restarts, the real-time engine keeps running.

### 2. High Concurrency via Go Goroutines
WebSockets require keeping a continuous, open connection between the browser and the server. Managing thousands of these open links can heavily drain a server's memory.
* **Why Go:** Runtimes like Node.js can struggle with memory when holding onto thousands of live connections. Go handles this using **Goroutines**—extremely lightweight internal threads.
* **The Benefit:** Each connection takes up almost zero RAM, allowing this small service to scale easily and handle a massive crowd of active users simultaneously.

### 3. Container Optimization (Multi-Stage Docker)
To make the application easy to deploy anywhere, the entire engine is containerised using a professional **Multi-Stage Dockerfile**.
* **Stage 1 (The Builder):** Uses a heavy Go environment to compile the raw code into a single, standalone binary file.
* **Stage 2 (The Runner):** Throws away the heavy compiler and copies *only* that finished binary into a tiny, secure Alpine Linux image.
* **The Result:** Because all the unnecessary source code and tools are left behind, the final production image is highly secure and ultra-lightweight (under **20MB**).

## 📂 Project Structure
```text
realtime-engine/
├── cmd/
│   └── server/
│       └── main.go       # App entry point, reads configuration & kicks off the server
├── internal/
│   └── websocket/
│       ├── client.go     # Manages individual reading/writing for each browser socket
│       └── hub.go        # The "control room" tracking who is connected and broadcasting data
├── .dockerignore         # Like a .gitignore, keeps local junk out of my Docker image
├── Dockerfile            # My optimized multi-stage build blueprint
├── go.mod                # Tracks my project dependencies (like Gorilla WebSocket & Redis drivers)
└── go.sum                # Security checksums for my dependencies
```

## 🛠️ How to Run and Test Locally
This project is configured completely through environment variables, keeping development settings separate from my actual code.

1. Start your local Redis Server\
Ensure Redis is installed on your Mac (installed via Homebrew):
```
brew services start redis
```

2. Build the Docker Image\
Compile the Go application into its optimised container setup.
```
docker build -t realtime-engine .
```

3. Run the Container\
Boot up the container, securely bridging its internal sandbox to your Mac's native Redis instance using `host.docker.internal`:

```
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e REDIS_ADDR=host.docker.internal:6379 \
  -e REDIS_CHANNEL=global_activity \
  realtime-engine
```

4. Connect a Test Client\
Open a new terminal tab and use `wscat` to mimic a frontend user connecting to the WebSocket layout:

```
npx wscat --connect ws://localhost:8080/ws
```

5. Broadcast a Message\
Open a third terminal tab, jump into the Redis CLI, and fire a message down the pipeline:
```
redis-cli
PUBLISH global_activity '{"message": "Hello from the database!"}'
```

Look back at your `wscat` window—you will see the JSON arrive instantly!