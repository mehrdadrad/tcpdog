package grpc

import "github.com/mehrdadrad/tcpdog/config"

type grpcConf struct {
	Server    string
	TLSConfig config.TLSConfig
}

func gRPCConfig(cfg map[string]interface{}) (*grpcConf, error) {
	// default config
	gCfg := &grpcConf{
		Server: "localhost:8085",
	}

	if err := config.Transform(cfg, gCfg); err != nil {
		return nil, err
	}

	return gCfg, nil
}
