package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

const (
	dialTimeout    = 10 * time.Second
	retryMax       = 5
	retryBackoff   = 2 * time.Second
	maxBackoff     = 10 * time.Second
	connectRetries = 30
)

type Broker struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	config   *Config
}

type Config struct {
	Brokers []string
	GroupID string
}

type Message struct {
	Topic string
	Key   string
	Value []byte
}

func NewBroker(cfg *Config) (*Broker, error) {
	conf := getKafkaConfig()

	if err := waitForKafka(cfg.Brokers); err != nil {
		return nil, fmt.Errorf("kafka connection failed: %w", err)
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, conf)
	if err != nil {
		return nil, fmt.Errorf("producer creation failed: %w", err)
	}

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, conf)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("consumer creation failed: %w", err)
	}

	return &Broker{producer, consumer, cfg}, nil
}

func getKafkaConfig() *sarama.Config {
	c := sarama.NewConfig()

	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.Retry.Max = retryMax
	c.Producer.Retry.Backoff = retryBackoff

	c.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Consumer.Offsets.AutoCommit.Enable = true

	c.Net.DialTimeout = dialTimeout
	c.Net.ReadTimeout = dialTimeout
	c.Net.WriteTimeout = dialTimeout
	c.Version = sarama.V2_0_0_0

	return c
}

func waitForKafka(brokers []string) error {
	conf := sarama.NewConfig()
	conf.Net.DialTimeout = dialTimeout

	var lastErr error
	backoff := time.Second

	for attempt := 1; attempt <= connectRetries; attempt++ {
		if client, err := sarama.NewClient(brokers, conf); err == nil {
			client.Close()
			return nil
		} else {
			lastErr = err
		}

		if attempt < connectRetries {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("connection failed after %d attempts: %w", connectRetries, lastErr)
}

func (b *Broker) Publish(ctx context.Context, msg *Message) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(msg.Key),
		Value: sarama.ByteEncoder(msg.Value),
	}

	if _, _, err := b.producer.SendMessage(kafkaMsg); err != nil {
		return fmt.Errorf("message publish failed: %w", err)
	}

	return nil
}

func (b *Broker) Subscribe(ctx context.Context, topic string, handler sarama.ConsumerGroupHandler) error {
	for {
		if err := b.consumer.Consume(ctx, []string{topic}, handler); err != nil {
			return fmt.Errorf("consume failed: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (b *Broker) Close() error {
	if err := b.producer.Close(); err != nil {
		return fmt.Errorf("producer close failed: %w", err)
	}
	if err := b.consumer.Close(); err != nil {
		return fmt.Errorf("consumer close failed: %w", err)
	}
	return nil
}
