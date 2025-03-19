package telemetry

import (
	"context"
	"fmt"
	"github.com/IndexStorm/common-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkrsc "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"time"
)

type OtelConfig struct {
	App                  config.AppInfo
	Trace                *config.TraceConfig
	Meter                *config.MeterConfig
	resource             *sdkrsc.Resource
	metricExportInterval time.Duration
}

type OtelOption interface {
	apply(cfg *OtelConfig)
}

func InitOpenTelemetryFromEnv(ctx context.Context, appConfig config.AppInfo, opts ...OtelOption) error {
	otelConfig := config.TryParseOpenTelemetry()
	if otelConfig == nil {
		return nil
	}
	return InitOpenTelemetry(ctx, OtelConfig{
		App:   appConfig,
		Trace: otelConfig.Trace,
		Meter: otelConfig.Meter,
	}, opts...)
}

func InitOpenTelemetry(ctx context.Context, cfg OtelConfig, opts ...OtelOption) error {
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	if cfg.resource == nil {
		resource, err := otelDefaultResource(ctx, cfg.App)
		if err != nil {
			return fmt.Errorf("init create default resource: %w", err)
		}
		cfg.resource = resource
	}
	if cfg.Trace != nil {
		if err := otelInitTraceProvider(ctx, cfg); err != nil {
			return fmt.Errorf("init trace provider: %w", err)
		}
	}
	if cfg.Meter != nil {
		if err := otelInitMeterProvider(ctx, cfg); err != nil {
			return fmt.Errorf("init meter provider: %w", err)
		}
	}
	return nil
}

func OtelWithResource(resource *sdkrsc.Resource) OtelOption {
	return &otelResourceOption{resource: resource}
}

type otelResourceOption struct {
	resource *sdkrsc.Resource
}

func (o *otelResourceOption) apply(cfg *OtelConfig) {
	cfg.resource = o.resource
}

func OtelWithMetricExportInterval(interval time.Duration) OtelOption {
	return &otelMetricExportIntervalOption{interval: interval}
}

type otelMetricExportIntervalOption struct {
	interval time.Duration
}

func (o *otelMetricExportIntervalOption) apply(cfg *OtelConfig) {
	cfg.metricExportInterval = o.interval
}

func otelDefaultResource(ctx context.Context, app config.AppInfo) (*sdkrsc.Resource, error) {
	return sdkrsc.New(
		ctx,
		sdkrsc.WithFromEnv(),
		sdkrsc.WithTelemetrySDK(),
		sdkrsc.WithProcess(),
		sdkrsc.WithOS(),
		sdkrsc.WithContainer(),
		sdkrsc.WithHost(),
		sdkrsc.WithAttributes(
			semconv.ServiceName(app.Service),
			semconv.ServiceNamespace(app.Namespace),
			semconv.ServiceVersion(app.Version),
			semconv.ServiceInstanceID(app.Instance),
		),
	)
}

func otelInitTraceProvider(ctx context.Context, cfg OtelConfig) error {
	var exporter *otlptrace.Exporter
	if cfg.Trace.Method == config.OpenTelemetryMethodHTTP {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Trace.Endpoint),
		}
		if cfg.Trace.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if cfg.Trace.Authorization != "" {
			opts = append(opts, otlptracehttp.WithHeaders(map[string]string{
				"Authorization": cfg.Trace.Authorization,
			}))
		}
		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("init http trace exporter: %w", err)
		}
		exporter = exp
	} else if cfg.Trace.Method == config.OpenTelemetryMethodGRPC {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Trace.Endpoint),
		}
		if cfg.Trace.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if cfg.Trace.Authorization != "" {
			opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
				"Authorization": cfg.Trace.Authorization,
			}))
		}
		exp, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("init grpc trace exporter: %w", err)
		}
		exporter = exp
	} else {
		return fmt.Errorf("invalid trace export method: %s", cfg.Trace.Method)
	}
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithIDGenerator(&lockFreeIdGenerator{}),
		sdktrace.WithResource(cfg.resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func otelInitMeterProvider(ctx context.Context, cfg OtelConfig) error {
	var exporter sdkmetric.Exporter
	if cfg.Meter.Method == config.OpenTelemetryMethodHTTP {
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(cfg.Meter.Endpoint),
		}
		if cfg.Meter.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		if cfg.Meter.Authorization != "" {
			opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
				"Authorization": cfg.Meter.Authorization,
			}))
		}
		exp, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("init http metric exporter: %w", err)
		}
		exporter = exp
	} else if cfg.Meter.Method == config.OpenTelemetryMethodGRPC {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.Meter.Endpoint),
		}
		if cfg.Meter.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		if cfg.Meter.Authorization != "" {
			opts = append(opts, otlpmetricgrpc.WithHeaders(map[string]string{
				"Authorization": cfg.Meter.Authorization,
			}))
		}
		exp, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("init grpc metric exporter: %w", err)
		}
		exporter = exp
	} else {
		return fmt.Errorf("invalid metric export method: %s", cfg.Meter.Method)
	}
	var readerOpts []sdkmetric.PeriodicReaderOption
	if cfg.metricExportInterval > 0 {
		readerOpts = append(readerOpts, sdkmetric.WithInterval(cfg.metricExportInterval))
	}
	periodicReader := sdkmetric.NewPeriodicReader(exporter, readerOpts...)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(cfg.resource),
		sdkmetric.WithReader(periodicReader),
	)
	otel.SetMeterProvider(mp)
	return nil
}
