package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"

	"github.com/mehrdadrad/tcpdog/config"
)

func TestIngress(t *testing.T) {
	// grpc
	ch := make(chan interface{}, 1)
	flow := config.Flow{Ingress: "grpc01"}
	conf := config.ServerConfig{
		Ingress: map[string]config.Ingress{
			"grpc01": {Type: "grpc"},
		},
	}

	conf.SetMockLogger("memory")

	ctx := conf.WithContext(context.Background())
	ingress(ctx, flow, ch)

	// kafka
	broker := sarama.NewMockBroker(t, 1)
	defer broker.Close()

	mockFetchResponse := sarama.NewMockFetchResponse(t, 1)
	mockFetchResponse.SetMessage("tcpdog", 0, int64(1234), sarama.StringEncoder(`{"v1":"k1"}`))

	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader("tcpdog", 0, broker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset("tcpdog", 0, sarama.OffsetOldest, 0).
			SetOffset("tcpdog", 0, sarama.OffsetNewest, 2345),
		"FetchRequest": mockFetchResponse,
	})

	flow = config.Flow{Ingress: "kafka01"}
	conf = config.ServerConfig{
		Ingress: map[string]config.Ingress{
			"kafka01": {Type: "kafka", Config: map[string]interface{}{
				"brokers": []string{broker.Addr()},
			}},
		},
	}

	conf.SetMockLogger("memory")

	ctx = conf.WithContext(context.Background())
	ingress(ctx, flow, ch)
}

func TestIngestion(t *testing.T) {
	// influxdb
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))

	ch := make(chan interface{}, 1)
	flow := config.Flow{Ingestion: "influxdb01"}
	conf := config.ServerConfig{
		Ingestion: map[string]config.Ingestion{
			"influxdb01": {Type: "influxdb", Config: map[string]interface{}{
				"url": server.URL,
			}},
		},
	}

	conf.SetMockLogger("memory")

	ctx := conf.WithContext(context.Background())
	ingestion(ctx, flow, ch)

	// elasticsearch
	flow = config.Flow{Ingestion: "elasticsearch01"}
	conf = config.ServerConfig{
		Ingestion: map[string]config.Ingestion{
			"elasticsearch01": {Type: "elasticsearch", Config: map[string]interface{}{
				"url": server.URL,
			}},
		},
	}

	conf.SetMockLogger("memory")

	ingestion(ctx, flow, ch)
}

func TestValidate(t *testing.T) {
	cfg := &config.ServerConfig{}

	assert.Error(t, validate(cfg))

	cfg.Flow = append(cfg.Flow, config.Flow{})
	assert.Error(t, validate(cfg))

	cfg.Flow[0].Ingestion = "ingestion01"
	assert.Error(t, validate(cfg))

	cfg.Flow[0].Ingress = "ingress01"
	assert.Error(t, validate(cfg))

	cfg.Flow[0].Serialization = "json"
	assert.Error(t, validate(cfg))

	cfg.Ingestion = map[string]config.Ingestion{"ingestion01": {}}
	assert.Error(t, validate(cfg))

	cfg.Ingress = map[string]config.Ingress{"ingress01": {}}
	assert.NoError(t, validate(cfg))
}
