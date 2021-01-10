package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ebpf"
)

func validation(cfg *config.Config) error {
	for i, tp := range cfg.Tracepoints {
		// fields validation
		err := validationFields(cfg, tp.Fields)
		if err != nil {
			return err
		}

		// tcpstatus validation
		s, err := ebpf.ValidateTCPStatus(tp.TCPState)
		if err != nil {
			return err
		}
		cfg.Tracepoints[i].TCPState = s

		// tracepoint
		err = ebpf.ValidateTracepoint(tp.Name)
		if err != nil {
			return err
		}

		// egress
		err = validationEgress(cfg)
		if err != nil {
			return err
		}

		// inet validation and default
		// TODO
	}

	return nil
}

func validationFields(cfg *config.Config, name string) error {
	if _, ok := cfg.Fields[name]; !ok {
		return fmt.Errorf("%s not exist", name)
	}

	for i, f := range cfg.Fields[name] {
		cf, err := ebpf.ValidateField(f.Name)
		if err != nil {
			return err
		}

		cfg.Fields[name][i].Name = cf
		cfg.Fields[name][i].Filter = strings.Replace(f.Filter, f.Name, cf, -1)
	}
	return nil
}

func validationEgress(cfg *config.Config) error {
	for _, tracepoint := range cfg.Tracepoints {
		if _, ok := cfg.Egress[tracepoint.Egress]; !ok {
			return fmt.Errorf("egress not found: %s", tracepoint.Egress)
		}
	}
	return nil
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
