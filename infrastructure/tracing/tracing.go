// Package tracing configures the global OpenTelemetry tracer provider (OTLP/HTTP
// export, W3C propagation). HTTP-perf metrics are delegated to the infra layer;
// only distributed tracing (intra-request breakdown) is instrumented in-app.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// Init configures the global tracer provider from the tracing config and the
// service identity, returning a shutdown func that flushes and stops it. When
// tracing is disabled it is a no-op: the global tracer stays the default no-op,
// so the app runs with no collector and zero tracing overhead.
func Init(
	ctx context.Context,
	cfg config.Tracing,
	serviceName, serviceVersion, environment string,
) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if !cfg.Enabled {
		return noop, nil
	}

	opts := []otlptracehttp.Option{}
	if cfg.OTLPEndpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.OTLPEndpoint), otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return noop, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	res := resource.NewSchemaless(
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
		semconv.DeploymentEnvironmentNameKey.String(environment),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}
