package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"realtime/internal/streams"
	"realtime/internal/types"
	ws "realtime/internal/websocket"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	log.Println("--- New Handshake Request Received ---")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}
	log.Println("Handshake Successful: Connection Upgraded")

	client := &ws.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan types.Message, 256),
	}

	client.Hub.Register <- client
	log.Println("Client sent to Hub registration")

	go client.WritePump()
	go client.ReadPump()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system defaults")
	}

	cfg := LoadConfig()
	hub := ws.NewHub()
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	go hub.Run()
	go listenToRedis(ctx, rdb, hub, cfg)

	// define /ws route, when someone visits, run handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(hub, w, r)
	})

	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		limit := int64(20)

		limit_str := r.URL.Query().Get("limit")

		if limit_str != "" {
			parsedLimit, err := strconv.ParseInt(limit_str, 10, 64)

			if err != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}

			limit = parsedLimit
		}

		messages, err2 := streams.ReadFromStream(r.Context(), rdb, cfg.RedisStream, limit)

		if err2 != nil {
			http.Error(w, "unable to read history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})

	// start the server
	log.Printf("Realtime engine starting on :%s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func listenToRedis(ctx context.Context, rdb *redis.Client, hub *ws.Hub, cfg *Config) {
	pubsub := rdb.Subscribe(ctx, cfg.RedisChannel)
	defer pubsub.Close()

	log.Printf("Listening to Redis on channel: %s", cfg.RedisChannel)

	for {
		msg, err := pubsub.ReceiveMessage(ctx)

		if err != nil {
			log.Printf("Redis error: %v", err)
			continue
		}

		var message types.Message
		err2 := json.Unmarshal([]byte(msg.Payload), &message)

		if err2 != nil {
			log.Printf("Redis unmarshalling error: %v", err2)
			continue
		}

		err3 := streams.SaveToStream(ctx, rdb, cfg.RedisStream, message)
		if err3 != nil {
			log.Printf("Redis stream saving error: %v", err3)
		}

		hub.Broadcasts <- message
	}
}

type Config struct {
	Port         string
	RedisAddr    string
	RedisChannel string
	RedisStream  string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisChannel := os.Getenv("REDIS_CHANNEL")
	if redisChannel == "" {
		redisChannel = "global_activity"
	}

	redisStream := os.Getenv("REDIS_STREAM")
	if redisStream == "" {
		redisStream = "message_history"
	}

	cfg := &Config{
		Port:         port,
		RedisAddr:    redisAddr,
		RedisChannel: redisChannel,
		RedisStream:  redisStream,
	}

	return cfg
}
