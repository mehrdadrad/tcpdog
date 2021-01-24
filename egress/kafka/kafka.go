package kafka

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/helper"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

type kafka struct {
	producer sarama.AsyncProducer
	bufpool  *sync.Pool
	dCh      chan *bytes.Buffer
	bCh      chan []byte
	jsonTail []byte
}

// Start starts producing the requested fields to kafka cluster.
func Start(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		cfg  = config.FromContext(ctx)
		kCfg = kafkaConfig(cfg.Egress[tp.Egress].Config)
	)

	sCfg, err := saramaConfig(kCfg)
	if err != nil {
		return err
	}

	err = sCfg.Validate()
	if err != nil {
		return err
	}

	k := kafka{
		bufpool: bufpool,
		dCh:     ch,
	}

	k.producer, err = sarama.NewAsyncProducer(kCfg.Brokers, sCfg)
	if err != nil {
		return err
	}

	k.hostname()

	switch kCfg.Serialization {
	case "spb":
		k.bCh = make(chan []byte, 1000)
		for i := 0; i < kCfg.Workers; i++ {
			go k.workerSPB(ctx, cfg.Fields[tp.Fields])
		}
		k.protobufLoop(ctx, kCfg.Topic)

	case "pb":
		k.bCh = make(chan []byte, 1000)
		for i := 0; i < kCfg.Workers; i++ {
			go k.workerPB(ctx, cfg.Fields[tp.Fields])
		}
		k.protobufLoop(ctx, kCfg.Topic)

	case "json":
		k.jsonLoop(ctx, kCfg.Topic)
	}

	return nil
}

// struct protobuf worker
func (k *kafka) workerSPB(ctx context.Context, fields []config.Field) {
	spb := helper.NewStructPB(fields)
	logger := config.FromContext(ctx).Logger()

	for {
		select {
		case buf := <-k.dCh:
			a := &pb.FieldsSPB{
				Fields: spb.Unmarshal(buf),
			}

			b, err := proto.Marshal(a)
			if err != nil {
				logger.Error("kafka", zap.Error(err))
			}

			k.bCh <- b
			k.bufpool.Put(buf)
		case <-ctx.Done():
			return
		}
	}

}

// protobuf worker
func (k *kafka) workerPB(ctx context.Context, fields []config.Field) {
	logger := config.FromContext(ctx).Logger()
	hostname, _ := os.Hostname()

	for {
		select {
		case buf := <-k.dCh:
			m := pb.Fields{}
			protojson.Unmarshal(buf.Bytes(), &m)
			m.Hostname = &hostname
			b, err := proto.Marshal(&m)
			if err != nil {
				logger.Error("kafka", zap.Error(err))
			}

			k.bCh <- b
			k.bufpool.Put(buf)
		case <-ctx.Done():
			return
		}
	}
}

func (k *kafka) jsonLoop(ctx context.Context, topic string) {
	logger := config.FromContext(ctx).Logger()

	go func() {
		for {
			select {
			case buf := <-k.dCh:
				select {
				case k.producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.ByteEncoder(k.addHostname(buf)),
				}:
				case err := <-k.producer.Errors():
					logger.Error("kafka", zap.Error(err))
				}

				k.bufpool.Put(buf)

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (k *kafka) protobufLoop(ctx context.Context, topic string) {
	logger := config.FromContext(ctx).Logger()

	go func() {
		for {
			select {
			//  protobuf (pb) and struct protobuf (spb) serializations
			case b := <-k.bCh:
				select {
				case k.producer.Input() <- &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.ByteEncoder(b),
				}:
				case err := <-k.producer.Errors():
					logger.Error("kafka", zap.Error(err))
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}

// addHostname adds hostname to encoded json and returns
// a fresh byte slice.
func (k *kafka) addHostname(buf *bytes.Buffer) []byte {
	b := make([]byte, 0, buf.Len()+len(k.jsonTail))
	b = append(b, buf.Bytes()...)
	b[len(b)-1] = ','
	b = append(b, k.jsonTail...)

	return b
}

func (k *kafka) hostname() {
	hostname, _ := os.Hostname()
	k.jsonTail = []byte(fmt.Sprintf("\"Hostname\":\"%s\"}", hostname))
}
