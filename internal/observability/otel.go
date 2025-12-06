package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitOTEL(serviceName string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	endpoint := os.Getenv("OTEL_COLLECTOR_URL")
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	// Modern OTEL Exporter Setup — NO client
	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}
