package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
)

type consumerGroup struct {
	group         sarama.ConsumerGroup
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

func newConsumerGroup(kCfg *Config) (*consumerGroup, error) {
	var err error
	sConfig := sarama.NewConfig()
	sConfig.ClientID = "tcpdog"
	sConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	sConfig.Version = sarama.V0_10_2_1

	group, err := sarama.NewConsumerGroup(kCfg.Brokers, "tcpdog", sConfig)
	if err != nil {
		return nil, err
	}

	return &consumerGroup{
		group: group,
	}, nil
}

// Start ...
func Start(ctx context.Context, name string, ser string, ch chan interface{}) {
	kCfg := influxConfig(config.FromContext(ctx).Ingress[name].Config)
	logger := config.FromContext(ctx).Logger()

	cg, err := newConsumerGroup(kCfg)
	if err != nil {
		logger.Fatal("kafka.consumer1", zap.Error(err))
	}

	cg.serialization = ser

	// error handling
	go func() {
		for err := range cg.group.Errors() {
			logger.Error("kafka.consumer2", zap.Error(err))
		}
	}()

	handler := handler{
		ch: make(chan []byte, 1),
	}

	// consumer group
	go func() {
		for {
			err := cg.group.Consume(ctx, []string{kCfg.Topic}, handler)
			if err != nil {
				logger.Fatal("kafka.consumer3", zap.Error(err))
			}
		}
	}()

	for i := 0; i < kCfg.Workers; i++ {
		go cg.worker(ctx, ch, handler.ch)
	}
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
			log.Println("marshal", err, string(b))
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
