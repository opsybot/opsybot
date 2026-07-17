package otel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/opsybot/opsybot/internal/config"
)

type Client struct {
	tracer *sdktrace.TracerProvider
	meter  *sdkmetric.MeterProvider
}

func From(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func New(cfg config.OTel, env config.Environment) (Client, func(), error) {
	if cfg.Endpoint == "" {
		return Client{}, func() {}, nil
	}

	ctx := context.Background()
	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.DeploymentEnvironmentNameKey.String(string(env)),
	))
	if err != nil {
		return Client{}, nil, fmt.Errorf("build otel resource: %w", err)
	}

	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}

	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return Client{}, nil, fmt.Errorf("new otel trace exporter: %w", err)
	}
	tracer := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)

	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		shutdown(tracer.Shutdown, cfg.ExportTimeout)
		return Client{}, nil, fmt.Errorf("new otel metric exporter: %w", err)
	}
	meter := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(cfg.MetricInterval))),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tracer)
	otel.SetMeterProvider(meter)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if err := runtime.Start(runtime.WithMeterProvider(meter)); err != nil {
		shutdown(meter.Shutdown, cfg.ExportTimeout)
		shutdown(tracer.Shutdown, cfg.ExportTimeout)
		return Client{}, nil, fmt.Errorf("start otel runtime metrics: %w", err)
	}

	c := Client{tracer: tracer, meter: meter}
	return c, func() {
		shutdown(c.meter.Shutdown, cfg.ExportTimeout)
		shutdown(c.tracer.Shutdown, cfg.ExportTimeout)
	}, nil
}

func shutdown(fn func(context.Context) error, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := fn(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		otel.Handle(err)
	}
}
