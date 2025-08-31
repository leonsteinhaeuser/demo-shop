package utils

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates a middleware that adds tracing to HTTP handlers
func TracingMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from incoming request
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start a new span
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Create a wrapped response writer to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Inject trace context into outgoing response headers
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Call the next handler with the traced context
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Add response attributes to the span
			span.SetAttributes(
				attribute.Int("http.status_code", wrappedWriter.statusCode),
			)

			// Set span status based on HTTP status code
			if wrappedWriter.statusCode >= 400 {
				span.RecordError(nil)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StartSpan is a helper function to start a span with common attributes
func StartSpan(ctx context.Context, operationName string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("demo-shop")
	return tracer.Start(ctx, operationName, trace.WithAttributes(attributes...))
}

// AddSpanEvent adds an event to the current span if one exists
func AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attributes...))
	}
}

// SetSpanError records an error on the current span if one exists
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// TracedHTTPClient creates an HTTP client that automatically propagates trace context
func TracedHTTPClient() *http.Client {
	return &http.Client{
		Transport: &tracedTransport{
			base: http.DefaultTransport,
		},
	}
}

// tracedTransport wraps an HTTP transport to inject trace context
type tracedTransport struct {
	base http.RoundTripper
}

func (t *tracedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new span for the outgoing HTTP request
	ctx := req.Context()
	tracer := otel.Tracer("demo-shop-http-client")

	ctx, span := tracer.Start(ctx, "HTTP "+req.Method,
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
			attribute.String("http.scheme", req.URL.Scheme),
			attribute.String("http.host", req.URL.Host),
			attribute.String("component", "http-client"),
		),
	)
	defer span.End()

	// Inject trace context into the outgoing request
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Update request with traced context
	req = req.WithContext(ctx)

	// Make the actual HTTP request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	// Add response attributes
	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	// Set span status based on HTTP status code
	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	return resp, nil
}

// InjectTraceHeaders manually injects trace context into HTTP headers
// Useful when you can't use TracedHTTPClient
func InjectTraceHeaders(ctx context.Context, headers http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractTraceContext extracts trace context from HTTP headers
// Useful for manual trace context extraction
func ExtractTraceContext(ctx context.Context, headers http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
}
