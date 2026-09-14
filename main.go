package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	redisotel "github.com/redis/go-redis/extra/redisotel-native/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const cacheMissesPerSecond = 10

func main() {
	ctx := context.Background()

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("otel-collector:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(time.Second))),
		sdkmetric.WithResource(resource.NewSchemaless(
			attribute.String("service.name", "redis-cache-miss-reproducer"),
		)),
	)
	otel.SetMeterProvider(provider)

	instrumentation := redisotel.GetObservabilityInstance()
	if err := instrumentation.Init(redisotel.NewConfig().
		WithEnabled(true).
		WithMetricGroups(redisotel.MetricGroupAll)); err != nil {
		log.Fatal(err)
	}
	defer instrumentation.Shutdown()
	defer func() { _ = provider.Shutdown(ctx) }()

	client := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	defer client.Close()

	key := fmt.Sprintf("missing:%d", time.Now().UnixNano())
	ticker := time.NewTicker(time.Second / cacheMissesPerSecond)
	defer ticker.Stop()

	for range ticker.C {
		err := client.Get(ctx, key).Err()
		if !errors.Is(err, redis.Nil) {
			log.Fatalf("GET %q: expected redis.Nil, got %v", key, err)
		}
	}
}
