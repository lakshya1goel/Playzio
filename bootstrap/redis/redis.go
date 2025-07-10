package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	pubsub *redis.PubSub
}

var RedisClient *Redis

func ConnectRedis(host, port, password string, db int) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
		PoolSize: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		fmt.Println("Redis connection failed with error: ", err)
		return err
	}

	RedisClient = &Redis{
		client: client,
	}

	fmt.Println("Redis connected successfully")
	return nil
}

func (r *Redis) Publish(channel string, message model.ChatMessage) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling message: ", err)
		return err
	}

	if err := r.client.Publish(context.Background(), channel, messageJSON).Err(); err != nil {
		fmt.Println("Error publishing message: ", err)
		return err
	}

	return nil
}

func (r *Redis) Subscribe(channel string, messageHandler func(model.ChatMessage)) error {
	pubsub := r.client.Subscribe(context.Background(), channel)
	r.pubsub = pubsub

	if _, err := pubsub.Receive(context.Background()); err != nil {
		fmt.Println("Error subscribing to channel: ", err)
		return err
	}

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var message model.ChatMessage
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				fmt.Println("Error unmarshalling message: ", err)
				continue
			}
			messageHandler(message)
		}
	}()

	return nil
}

func (r *Redis) Unsubscribe(channel string) error {
	if r.pubsub != nil {
		if err := r.pubsub.Unsubscribe(context.Background(), channel); err != nil {
			fmt.Println("Error unsubscribing from channel: ", err)
			return err
		}
	}
	return nil
}

func (r *Redis) Close() error {
	if r.pubsub != nil {
		if err := r.pubsub.Close(); err != nil {
			fmt.Println("Error closing pubsub: ", err)
			return err
		}
	}
	return nil
}
