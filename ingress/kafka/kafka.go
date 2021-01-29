package kafka

import (
	"context"
	"encoding/json"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/helper"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

type consumerGroup struct {
	group         sarama.ConsumerGroup
	logger        *zap.Logger
	serialization string
}

type handler struct {
	ch chan []byte
}

func (h handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.ch <- message.Value
		session.MarkMessage(message, "")
	}
	return nil
}

func newConsumerGroup(logger *zap.Logger, kCfg *Config) (*consumerGroup, error) {
	var err error

	sConfig, err := saramaConfig(kCfg)
	if err != nil {
		return nil, err
	}

	group, err := sarama.NewConsumerGroup(kCfg.Brokers, "tcpdog", sConfig)
	if err != nil {
		return nil, err
	}

	return &consumerGroup{
		group:  group,
		logger: logger,
	}, nil
}

// Start starts a consumer group
func Start(ctx context.Context, name string, ser string, ch chan interface{}) error {
	kCfg := kafkaConfig(config.FromContextServer(ctx).Ingress[name].Config)
	logger := config.FromContextServer(ctx).Logger()

	cg, err := newConsumerGroup(logger, kCfg)
	if err != nil {
		return err
	}

	cg.serialization = ser

	// error handling
	go func() {
		for err := range cg.group.Errors() {
			logger.Error("kafka", zap.Error(err))
		}
	}()

	handler := handler{
		ch: make(chan []byte, 1),
	}

	// consumer group
	go func() {
		backoff := helper.NewBackoff(logger)

		for {
			backoff.Next()

			err := cg.group.Consume(ctx, []string{kCfg.Topic}, handler)
			if err != nil {
				logger.Error("kafka", zap.Error(err))
			} else {
				logger.Warn("kafka", zap.String("msg", "consumer group has been terminated"))
				cg.consumerGroupCleanup()
				return
			}
		}
	}()

	for i := 0; i < kCfg.Workers; i++ {
		go cg.worker(ctx, ch, handler.ch)
	}

	return nil
}

func (k *consumerGroup) consumerGroupCleanup() {
	k.group.Close()
}

func (k *consumerGroup) worker(ctx context.Context, ch chan interface{}, bCh chan []byte) {
	unmarshal := getUnmarshal(k.serialization)

	for {
		b := <-bCh
		i, err := unmarshal(b)
		if err != nil {
			k.logger.Error("kafka", zap.String("event", "marshal"), zap.Error(err))
			continue
		}

		ch <- i
	}
}

func getUnmarshal(ser string) func(b []byte) (interface{}, error) {
	switch ser {
	case "json":
		return func(b []byte) (interface{}, error) {
			m := map[string]interface{}{}
			err := json.Unmarshal(b, &m)
			return m, err
		}
	case "spb":
		return func(b []byte) (interface{}, error) {
			p := pb.FieldsSPB{}
			err := proto.Unmarshal(b, &p)
			return &p, err
		}
	case "pb":
		return func(b []byte) (interface{}, error) {
			p := pb.Fields{}
			err := proto.Unmarshal(b, &p)
			return &p, err
		}
	}

	return nil
}
