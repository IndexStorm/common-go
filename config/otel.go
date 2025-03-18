package config

import (
	"github.com/caarlos0/env/v11"
	"time"
)

type OpenTelemetryMethod string

const (
	OpenTelemetryMethodHTTP OpenTelemetryMethod = "HTTP"
	OpenTelemetryMethodGRPC OpenTelemetryMethod = "GRPC"
)

type OpenTelemetry struct {
	Trace *TraceConfig `envPrefix:"TRACE_"`
	Meter *MeterConfig `envPrefix:"METER_"`
}

type TraceConfig struct {
	Endpoint      string              `env:"ENDPOINT,notEmpty,unset"`
	Method        OpenTelemetryMethod `env:"METHOD,notEmpty"`
	Insecure      bool                `env:"INSECURE"`
	Authorization string              `env:"AUTHORIZATION,unset"`
}

type MeterConfig struct {
	Endpoint      string              `env:"ENDPOINT,notEmpty,unset"`
	Method        OpenTelemetryMethod `env:"METHOD,notEmpty"`
	Insecure      bool                `env:"INSECURE"`
	Authorization string              `env:"AUTHORIZATION,unset"`
	Interval      time.Duration       `env:"INTERVAL" envDefault:"1m"`
}

func TryParseOpenTelemetry() *OpenTelemetry {
	var telemetryCfg OpenTelemetry
	if traceCfg, err := env.ParseAsWithOptions[TraceConfig](env.Options{Prefix: "OTEL_TRACE_"}); err == nil {
		telemetryCfg.Trace = &traceCfg
	}
	if meterCfg, err := env.ParseAsWithOptions[MeterConfig](env.Options{Prefix: "OTEL_METER_"}); err == nil {
		telemetryCfg.Meter = &meterCfg
	}
	if telemetryCfg.Trace != nil || telemetryCfg.Meter != nil {
		return &telemetryCfg
	}
	return nil
}
