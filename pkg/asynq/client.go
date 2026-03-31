package asynq

import (
	"crypto/tls"
	"time"

	hibikenasynq "github.com/hibiken/asynq"
)

type AsynqClient struct {
	client *hibikenasynq.Client
}

func NewClient(redisHost, redisUser, redisPassword string, tlsConfig *tls.Config) *AsynqClient {
	client := hibikenasynq.NewClient(hibikenasynq.RedisClientOpt{
		Addr:         redisHost,
		Username:     redisUser,
		Password:     redisPassword,
		TLSConfig:    tlsConfig,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	return &AsynqClient{client: client}
}

func (c *AsynqClient) GetClient() *hibikenasynq.Client {
	return c.client
}

func (c *AsynqClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}

	return nil
}
