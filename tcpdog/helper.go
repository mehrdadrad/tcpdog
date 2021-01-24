package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ebpf"
)

func validate(cfg *config.Config) error {
	for i, tp := range cfg.Tracepoints {
		// fields validation
		err := validateFields(cfg, tp.Fields)
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

		// egress, inet and sample
		err = validateMix(cfg, tp)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateFields(cfg *config.Config, name string) error {
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

func validateMix(cfg *config.Config, tp config.Tracepoint) error {
	if _, ok := cfg.Egress[tp.Egress]; !ok {
		return fmt.Errorf("egress not found: %s", tp.Egress)
	}

	for _, inet := range tp.INet {
		if inet != 4 && inet != 6 {
			return fmt.Errorf("wrong inet version (%s) inet:%d", tp.Name, inet)
		}
	}

	if tp.Sample < 0 {
		return fmt.Errorf("negative sample (%s) sample:%d", tp.Name, tp.Sample)
	}

	return nil
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
