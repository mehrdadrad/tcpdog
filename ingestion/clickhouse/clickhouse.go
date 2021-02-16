package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	chgo "github.com/ClickHouse/clickhouse-go"
	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/helper"
	"github.com/mehrdadrad/tcpdog/geo"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

type clickhouse struct {
	geo           geo.Geoer
	cfg           *chConfig
	serialization string

	vFields reflect.Value
}

// Start starts ingestion data to clickhouse
func Start(ctx context.Context, name string, ser string, ch chan interface{}) error {
	var g geo.Geoer

	cfg := config.FromContextServer(ctx)

	cCfg, err := clickhouseConfig(cfg.Ingestion[name].Config)
	if err != nil {
		return err
	}

	connect, err := sql.Open("clickhouse", cCfg.DSName)
	if err != nil {
		return err
	}

	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*chgo.Exception); ok {
			return fmt.Errorf("[%d] %s %s", exception.Code, exception.Message, exception.StackTrace)
		}
		return err
	}

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	c := clickhouse{geo: g, cfg: cCfg, serialization: ser, vFields: reflect.ValueOf(&pb.Fields{}).Elem()}
	iCh := make(chan []interface{}, 1000)

	for i := 0; i < c.cfg.Workers; i++ {
		go c.iWorker(ctx, ch, iCh)
	}

	for i := 0; i < c.cfg.Connections; i++ {
		go c.ingest(ctx, connect, iCh)
	}

	return nil
}

func (c *clickhouse) iWorker(ctx context.Context, ch chan interface{}, iCh chan []interface{}) {
	fn := c.getSliceIfMaker()
	logger := config.FromContextServer(ctx).Logger()

	for {
		select {
		case data := <-ch:
			s, err := fn(data)
			if err != nil {
				logger.Error("clickhouse", zap.Error(err))
				continue
			}

			iCh <- s
		case <-ctx.Done():
			return
		}
	}
}

func (c *clickhouse) ingest(ctx context.Context, connect *sql.DB, iCh chan []interface{}) {
	query := c.getQuery()
	logger := config.FromContextServer(ctx).Logger()
	timer := time.NewTimer(time.Second * 10)
	backoff := helper.NewBackoff(logger)
	interval := time.Second * time.Duration(c.cfg.FlushInterval)
	counter := 0
	timeoutCounter := 0

OUTERLOOP:
	for {

		tx, err := connect.Begin()
		if err != nil {
			logger.Error("clickhouse-1", zap.Error(err))
			backoff.Next()
			continue
		}

		stmt, err := tx.Prepare(query)
		if err != nil {
			logger.Error("clickhouse-2", zap.Error(err))
			backoff.Next()
			continue
		}

		counter = 0
		timeoutCounter = 0
		timer.Reset(interval)

	INNERLOOP:
		for {
			select {
			case fields := <-iCh:
				_, err := stmt.ExecContext(ctx, fields...)
				if err != nil {
					logger.Error("clickhouse-3", zap.Error(err))
				}

				counter++
				if counter >= c.cfg.BatchSize {
					break INNERLOOP
				}
			case <-timer.C:
				if counter > 0 {
					break INNERLOOP
				}
				timer.Reset(interval)
				if timeoutCounter++; timeoutCounter >= (300/c.cfg.FlushInterval)-1 {
					continue OUTERLOOP
				}
			case <-ctx.Done():
				tx.Commit()
				return
			}
		}

		if err := tx.Commit(); err != nil {
			logger.Error("clickhouse", zap.Error(err))
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}
}

func (c *clickhouse) JSON(fi interface{}) ([]interface{}, error) {
	f := fi.(map[string]interface{})

	if c.geo != nil {
		if gv, ok := f[c.cfg.GeoField]; ok {
			for k, v := range c.geo.Get(gv.(string)) {
				f[k] = v
			}
		}
	}

	a := make([]interface{}, len(c.cfg.Fields))

	for i, name := range c.cfg.Fields {
		switch c.vFields.FieldByName(name).Type().Elem().Kind() {
		case reflect.Uint32:
			a[i] = uint32(f[name].(float64))
		case reflect.Uint64:
			a[i] = uint64(f[name].(float64))
		case reflect.String:
			a[i] = f[name]
		}
	}

	return a, nil
}

func (c *clickhouse) PB(fi interface{}) ([]interface{}, error) {
	a := make([]interface{}, len(c.cfg.Fields))
	v := reflect.ValueOf(fi.(*pb.Fields)).Elem()

	geoKV := map[string]string{}
	if c.geo != nil {
		gv := v.FieldByName(c.cfg.GeoField)
		if gv.IsValid() {
			geoKV = c.geo.Get(gv.Elem().String())
		}
	}

	for i, name := range c.cfg.Fields {
		if v.FieldByName(name).Pointer() != 0 {
			switch v.FieldByName(name).Type().Elem().Kind() {
			case reflect.Uint32:
				a[i] = uint32(v.FieldByName(name).Elem().Uint())
			case reflect.Uint64:
				a[i] = v.FieldByName(name).Elem().Uint()
			case reflect.String:
				a[i] = v.FieldByName(name).Elem().String()
			}
		} else if g, ok := geoKV[name]; ok {
			a[i] = g
		}
	}

	return a, nil
}

func (c *clickhouse) SPB(fi interface{}) ([]interface{}, error) {
	f := fi.(*pb.FieldsSPB)
	a := make([]interface{}, len(c.cfg.Fields))

	geoKV := map[string]string{}
	if c.geo != nil {
		gv, ok := f.Fields.Fields[c.cfg.GeoField]
		if ok {
			geoKV = c.geo.Get(gv.GetStringValue())
		}
	}

	for i, name := range c.cfg.Fields {
		switch c.vFields.FieldByName(name).Type().Elem().Kind() {
		case reflect.Uint32:
			a[i] = uint32(f.Fields.Fields[name].GetNumberValue())
		case reflect.Uint64:
			a[i] = uint64(f.Fields.Fields[name].GetNumberValue())
		case reflect.String:
			if value, ok := f.Fields.Fields[name]; ok {
				a[i] = value.GetStringValue()
			} else {
				a[i] = geoKV[name]
			}
		}
	}

	return a, nil
}

func (c *clickhouse) getQuery() string {
	qm := make([]string, len(c.cfg.Columns))
	for i := range c.cfg.Columns {
		qm[i] = "?"
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		c.cfg.Table, strings.Join(c.cfg.Columns, ","),
		strings.Join(qm, ","))
}

func (c *clickhouse) getSliceIfMaker() func(fi interface{}) ([]interface{}, error) {
	switch c.serialization {
	case "json":
		return c.JSON
	case "spb":
		return c.SPB
	case "pb":
		return c.PB
	}

	return nil
}
