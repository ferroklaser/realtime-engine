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
