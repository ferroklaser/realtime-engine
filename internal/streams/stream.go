package streams

import (
	"context"
	"encoding/json"
	"log"
	"realtime/internal/types"

	"github.com/redis/go-redis/v9"
)

func SaveToStream(ctx context.Context, rdb *redis.Client, stream string, msg types.Message) error {
	jsonData, err := json.Marshal(msg)

	if err != nil {
		log.Println("Error marshalling for stream")
		return err
	}

	_, err2 := rdb.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{
				"payload": string(jsonData),
			},
		}).Result()

	if err2 != nil {
		log.Println("Error adding to stream")
		return err2
	}

	return nil
}

func ReadFromStream(ctx context.Context, rdb *redis.Client, stream string, limit int64) ([]types.Message, error) {
	msgs, err := rdb.XRevRangeN(
		ctx,
		stream,
		"+",
		"-",
		limit,
	).Result()

	if err != nil {
		log.Println("Error reading stream")
		return nil, err
	}

	var messages []types.Message

	for _, msg := range msgs {

		payload, ok := msg.Values["payload"].(string)

		if !ok {
			log.Printf("Redis unable to cast payload")
			continue
		}

		var message types.Message
		err2 := json.Unmarshal([]byte(payload), &message)

		if err2 != nil {
			log.Printf("Redis unmarshalling error: %v", err2)
			continue
		}

		message.ID = msg.ID
		messages = append(messages, message)
	}

	return messages, nil
}
