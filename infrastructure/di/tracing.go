package di

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

// GetTracerProvider returns the memoized OpenTelemetry tracer provider and
// registers it (with W3C propagation) as the global provider. The caller owns
// its lifecycle via provider.Shutdown (flush). When tracing is disabled the
// provider never samples — zero spans, no exporter — so there is no collector
// dependency and effectively no overhead.
func (c *Container) GetTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	if c.tracerProvider != nil {
		return c.tracerProvider, nil
	}

	res := resource.NewSchemaless(
		semconv.ServiceName(config.AppName),
		semconv.ServiceVersion(config.AppVersion),
		semconv.DeploymentEnvironmentNameKey.String(c.config.Env.Type),
	)

	var provider *sdktrace.TracerProvider

	if c.config.Tracing.Enabled {
		opts := []otlptracehttp.Option{}
		if c.config.Tracing.OTLPEndpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(c.config.Tracing.OTLPEndpoint), otlptracehttp.WithInsecure())
		}

		exporter, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}

		provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.config.Tracing.SampleRatio))),
		)
	} else {
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
	}

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	c.tracerProvider = provider

	return provider, nil
}
