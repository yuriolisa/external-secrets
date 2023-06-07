/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func newOtelResource(ctx context.Context) *resource.Resource {
	otelRes, err := resource.New(
		ctx, resource.WithAttributes(semconv.ServiceName("externalsecret_tracing")),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		fmt.Errorf("failed to create the otel resource: %w", err)
	}
	return otelRes
}

func NewTraceClient(ctx context.Context) (*sdktrace.TracerProvider, error) {
	//var client otlptrace.Client
	client := otlptracehttp.NewClient()

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		fmt.Errorf("failed to start the trace exporter: %w", err)
	}

	spanProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithResource(newOtelResource(ctx)),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func ShutDownController(ctx context.Context, traceProvider *sdktrace.TracerProvider) {
	// pushes any last exports to the receiver
	if err := traceProvider.Shutdown(ctx); err != nil {
		otel.Handle(err)
	}
}
