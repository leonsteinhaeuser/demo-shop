# OpenTelemetry Tracing Setup

This document explains how to use the OTLP (OpenTelemetry Protocol) tracer setup in the demo-shop project.

## Overview

The project includes a comprehensive tracing setup that supports both HTTP and gRPC OTLP exporters. The tracing implementation is located in `internal/utils/tracer.go` and includes helper functions for easy integration.

## Features

- Support for both HTTP and gRPC OTLP protocols
- Configurable service metadata (name, version, environment)
- Proper resource configuration with semantic conventions
- Graceful shutdown handling
- HTTP middleware for automatic request tracing
- Helper functions for manual span creation and management

## Configuration

The tracing setup uses environment variables for configuration:

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `TRACING_ENABLED` | Enable/disable tracing | `false` | `true` |
| `TRACING_ENDPOINT` | OTLP endpoint URL | `http://localhost:4318` | `http://jaeger:4318` |
| `TRACING_PROTOCOL` | Protocol to use (http/grpc) | `http` | `grpc` |
| `TRACING_INSECURE` | Use insecure connection | `true` | `false` |
| `ENVIRONMENT` | Deployment environment | `development` | `production` |

### OTLP Endpoints

**HTTP Protocol:**

- Jaeger: `http://localhost:4318` (default OTLP/HTTP port)
- OTEL Collector: `http://localhost:4318/v1/traces`

**gRPC Protocol:**

- Jaeger: `localhost:4317` (default OTLP/gRPC port)
- OTEL Collector: `localhost:4317`

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/leonsteinhaeuser/demo-shop/internal/utils"
)

func main() {
    ctx := context.Background()

    // Configure tracer
    config := utils.TracerConfig{
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        Endpoint:       "http://localhost:4318",
        Insecure:       true,
        Headers:        map[string]string{
            "Authorization": "Bearer token",
        },
    }

    // Create tracer (supports "http" or "grpc" protocol)
    tracer, shutdown, err := utils.NewTracer(ctx, "http", config)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err := shutdown(ctx); err != nil {
            log.Printf("Failed to shutdown tracer: %v", err)
        }
    }()

    // Use tracer for manual instrumentation
    ctx, span := tracer.Start(ctx, "my-operation")
    defer span.End()

    // Your application logic here
}
```

### HTTP Middleware

Use the provided middleware to automatically trace HTTP requests:

```go
package main

import (
    "net/http"
    
    "github.com/leonsteinhaeuser/demo-shop/internal/utils"
)

func main() {
    // Setup tracer first (see basic setup above)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)
    
    // Wrap with tracing middleware
    tracedHandler := utils.TracingMiddleware("my-service")(mux)
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: tracedHandler,
    }
    
    server.ListenAndServe()
}
```

### Manual Span Creation

Use helper functions for manual span management:

```go
package main

import (
    "context"
    "errors"
    
    "github.com/leonsteinhaeuser/demo-shop/internal/utils"
    "go.opentelemetry.io/otel/attribute"
)

func processOrder(ctx context.Context, orderID string) error {
    // Start a new span
    ctx, span := utils.StartSpan(ctx, "process-order",
        attribute.String("order.id", orderID),
        attribute.String("operation", "payment"),
    )
    defer span.End()
    
    // Add events to the span
    utils.AddSpanEvent(ctx, "order-validation-started")
    
    // Simulate some work
    if err := validateOrder(ctx, orderID); err != nil {
        // Record error on span
        utils.SetSpanError(ctx, err)
        return err
    }
    
    utils.AddSpanEvent(ctx, "order-processing-completed")
    return nil
}
```

## Integration Example

Here's how the gateway service integrates tracing:

```go
// cmd/gateway/main.go
func main() {
    ctx := context.Background()
    
    // Setup tracing if enabled
    var tracerShutdown func(context.Context) error
    if envTracingEnabled == "true" {
        config := utils.TracerConfig{
            ServiceName:    "demo-shop-gateway",
            ServiceVersion: version,
            Environment:    envEnvironment,
            Endpoint:       envTracingEndpoint,
            Insecure:       envTracingInsecure == "true",
        }

        _, shutdown, err := utils.NewTracer(ctx, envTracingProtocol, config)
        if err != nil {
            slog.Error("Failed to create tracer", "error", err)
        } else {
            tracerShutdown = shutdown
            slog.Info("Tracing enabled", "endpoint", envTracingEndpoint)
        }
    }
    
    // ... server setup ...
    
    // Graceful shutdown
    defer func() {
        if tracerShutdown != nil {
            if err := tracerShutdown(ctx); err != nil {
                log.Printf("Failed to shutdown tracer: %v", err)
            }
        }
    }()
}
```

## Testing with Jaeger

To test the tracing setup locally with Jaeger:

1. **Start Jaeger:**

   ```bash
   docker run -d --name jaeger \
     -p 16686:16686 \
     -p 4317:4317 \
     -p 4318:4318 \
     jaegertracing/all-in-one:latest
   ```

2. **Configure your service:**

   ```bash
   export TRACING_ENABLED=true
   export TRACING_ENDPOINT=http://localhost:4318
   export TRACING_PROTOCOL=http
   export TRACING_INSECURE=true
   ```

3. **Start your service and generate some traffic**

4. **View traces in Jaeger UI:**
   Open <http://localhost:16686> in your browser

## Production Considerations

### Security

- Set `TRACING_INSECURE=false` in production
- Use proper authentication headers if required
- Consider network security between services and OTLP endpoint

### Performance

- The default sampler is `AlwaysSample()` - consider using probability sampling in high-traffic production environments
- Configure appropriate batch sizes for the exporter
- Monitor the overhead of tracing on your application performance

### Reliability

- Implement proper error handling for tracer setup failures
- Consider using a local OTEL Collector for better reliability
- Set appropriate timeouts for trace export

## Troubleshooting

### Common Issues

1. **Traces not appearing:**
   - Verify `TRACING_ENABLED=true`
   - Check endpoint URL and port
   - Ensure the OTLP receiver is running and accessible

2. **Connection errors:**
   - Verify the protocol (http vs grpc) matches your receiver
   - Check firewall/network configuration
   - Verify TLS/insecure settings

3. **Build errors:**
   - Run `go mod tidy` to ensure all dependencies are available
   - Check Go version compatibility (requires Go 1.21+)

### Debug Mode

Enable debug logging to troubleshoot tracing issues:

```go
import "go.opentelemetry.io/otel/log/global"

// Add this before creating the tracer
global.SetLogger(logr.New(&debugLogger{}))
```

## Dependencies

The tracing setup requires these OpenTelemetry packages:

```go
go.opentelemetry.io/otel v1.38.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
go.opentelemetry.io/otel/sdk v1.38.0
go.opentelemetry.io/otel/semconv/v1.17.0 v1.17.0
```

Run `go mod tidy` to install all required dependencies.
