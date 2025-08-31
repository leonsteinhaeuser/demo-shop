package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/leonsteinhaeuser/demo-shop/internal/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	DefaultTracer trace.Tracer = otel.Tracer("demo-shop-default")
)

// TracerConfig holds configuration for the tracer setup
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	Insecure       bool
	Headers        map[string]string
	TracerProtocol string
}

func TraceConfigFromEnv() TracerConfig {
	return TracerConfig{
		ServiceName:    env.StringEnvOrDefault("SERVICE_NAME", "demo-shop"),
		ServiceVersion: env.StringEnvOrDefault("TRACING_SERVICE_VERSION", "1.0.0"),
		Endpoint:       env.StringEnvOrDefault("TRACING_ENDPOINT", "http://localhost:4318"),
		Insecure:       env.BoolEnvOrDefault("TRACING_INSECURE", true),
		Headers:        env.MapEnvOrDefault("TRACING_HEADERS", nil),
		TracerProtocol: env.StringEnvOrDefault("TRACING_PROTOCOL", "grpc"),
	}
}

// NewTracerGrpc creates a new tracer with OTLP gRPC exporter
func NewTracerGrpc(ctx context.Context, config TracerConfig) (trace.Tracer, func(context.Context) error, error) {
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP gRPC exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
	}

	if config.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if len(config.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP gRPC exporter: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := otel.Tracer(config.ServiceName)

	// Set as default tracer
	DefaultTracer = tracer

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}

	return tracer, shutdown, nil
}

// NewTracerHttp creates a new tracer with OTLP HTTP exporter
func NewTracerHttp(ctx context.Context, config TracerConfig) (trace.Tracer, func(context.Context) error, error) {
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
	}

	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(config.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(config.Headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := otel.Tracer(config.ServiceName)

	// Set as default tracer
	DefaultTracer = tracer

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}

	return tracer, shutdown, nil
}

// NewTracer creates a new tracer based on protocol from config
func NewTracer(ctx context.Context, config TracerConfig) (trace.Tracer, func(context.Context) error, error) {
	switch config.TracerProtocol {
	case "grpc":
		return NewTracerGrpc(ctx, config)
	case "http":
		return NewTracerHttp(ctx, config)
	default:
		return nil, nil, fmt.Errorf("unsupported protocol: %s, supported protocols are 'http' and 'grpc'", config.TracerProtocol)
	}
}

// SpanFromContext creates a new child span from context
// The "DefaultTracer" is used to create the span
func SpanFromContext(ctx context.Context, name string) (context.Context, trace.Span) {
	return DefaultTracer.Start(ctx, name)
}
