package otelutil

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"
)

// InitTracer configure tracer
func InitTracer(enabled bool, agentAddress string, serviceNameKey string) (context.CancelFunc, error) {
	lg := logrus.WithField("component", "utils.otel")
	nilCf := func() {}

	if !enabled {
		lg.Debug("tracer disabled")
		return nilCf, nil
	}

	if agentAddress == "" {
		lg.Debug("use stdout tracer")

		exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
		if err != nil {
			lg.WithError(err).Error("failed to create exporter")
			return nilCf, fmt.Errorf("failed to create exporter: %w", err)
		}

		bsp := sdktrace.NewBatchSpanProcessor(exporter)
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(bsp),
		)

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		return nilCf, nil
	}

	lg.Debug("setting up agent connection")
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` address. Otherwise, replace `localhost` with the
	// address of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(agentAddress),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)

	exporter, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		lg.WithError(err).Error("failed to create exporter")
		return nilCf, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceNameKey),
		),
	)
	if err != nil {
		lg.WithError(err).Error("failed to create resource")
		return nilCf, fmt.Errorf("failed to create resource: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exporter,
		),
		controller.WithExporter(exporter),
		controller.WithCollectPeriod(2*time.Second),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)
	global.SetMeterProvider(cont.MeterProvider())

	err = cont.Start(ctx)
	if err != nil {
		return nilCf, fmt.Errorf("failed to start controller: %w", err)
	}

	return func() {
		lg.Debug("stopping tracer")

		err := cont.Stop(ctx)
		if err != nil {
			lg.WithError(err).Fatal("failed to Stop controller")
		}

		err = tp.Shutdown(ctx)
		if err != nil {
			lg.WithError(err).Fatal("failed to Shutdown provider")
		}
	}, nil
}
