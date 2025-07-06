package util

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	prometheusexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type TelemetryManager struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Logger         *slog.Logger
	Metrics        *BubbleBankMetrics
	cleanup        func()
}

type BubbleBankMetrics struct {
	// HTTP metrics
	HTTPDuration metric.Float64Histogram
	HTTPRequests metric.Int64Counter
	HTTPErrors   metric.Int64Counter

	// Database metrics
	DBConnections metric.Int64UpDownCounter
	DBOperations  metric.Int64Counter
	DBDuration    metric.Float64Histogram
	DBErrors      metric.Int64Counter

	// Business metrics
	AccountsCreated metric.Int64Counter
	TransfersTotal  metric.Int64Counter
	TransferAmount  metric.Float64Histogram
	BalanceChanges  metric.Float64Histogram
}

func InitTelemetry(ctx context.Context, config Config) (*TelemetryManager, error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.OtelServiceName),
			semconv.ServiceVersion(config.OtelServiceVersion),
			semconv.DeploymentEnvironment(config.OtelEnvironment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracing
	tracerProvider, err := initTracing(ctx, config, res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics
	meterProvider, err := initMetrics(ctx, config, res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Set global providers
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Initialize structured logger
	logger := NewStructuredLogger(config)

	// Initialize business metrics
	metrics, err := initBusinessMetrics(meterProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize business metrics: %w", err)
	}

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+config.MetricsPort, nil); err != nil {
			logger.Error("metrics server failed", "error", err)
		}
	}()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown tracer provider", "error", err)
		}

		if err := meterProvider.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown meter provider", "error", err)
		}
	}

	return &TelemetryManager{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		Logger:         logger,
		Metrics:        metrics,
		cleanup:        cleanup,
	}, nil
}

func (tm *TelemetryManager) Shutdown() {
	tm.cleanup()
}

func initTracing(ctx context.Context, config Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// Create OTLP trace exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(config.OtelExporterOtlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return tracerProvider, nil
}

func initMetrics(ctx context.Context, config Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// Create Prometheus exporter
	exporter, err := prometheusexporter.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	return meterProvider, nil
}

func initBusinessMetrics(meterProvider *sdkmetric.MeterProvider) (*BubbleBankMetrics, error) {
	meter := meterProvider.Meter("bubblebank")

	httpDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	httpRequests, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	httpErrors, err := meter.Int64Counter(
		"http_errors_total",
		metric.WithDescription("Total HTTP errors"),
	)
	if err != nil {
		return nil, err
	}

	dbConnections, err := meter.Int64UpDownCounter(
		"db_connections_active",
		metric.WithDescription("Active database connections"),
	)
	if err != nil {
		return nil, err
	}

	dbOperations, err := meter.Int64Counter(
		"db_operations_total",
		metric.WithDescription("Total database operations"),
	)
	if err != nil {
		return nil, err
	}

	dbDuration, err := meter.Float64Histogram(
		"db_operation_duration_seconds",
		metric.WithDescription("Database operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	dbErrors, err := meter.Int64Counter(
		"db_errors_total",
		metric.WithDescription("Total database errors"),
	)
	if err != nil {
		return nil, err
	}

	accountsCreated, err := meter.Int64Counter(
		"accounts_created_total",
		metric.WithDescription("Total accounts created"),
	)
	if err != nil {
		return nil, err
	}

	transfersTotal, err := meter.Int64Counter(
		"transfers_total",
		metric.WithDescription("Total transfers"),
	)
	if err != nil {
		return nil, err
	}

	transferAmount, err := meter.Float64Histogram(
		"transfer_amount_histogram",
		metric.WithDescription("Transfer amount distribution"),
		metric.WithUnit("currency"),
	)
	if err != nil {
		return nil, err
	}

	balanceChanges, err := meter.Float64Histogram(
		"balance_changes_histogram",
		metric.WithDescription("Account balance changes distribution"),
		metric.WithUnit("currency"),
	)
	if err != nil {
		return nil, err
	}

	return &BubbleBankMetrics{
		HTTPDuration:    httpDuration,
		HTTPRequests:    httpRequests,
		HTTPErrors:      httpErrors,
		DBConnections:   dbConnections,
		DBOperations:    dbOperations,
		DBDuration:      dbDuration,
		DBErrors:        dbErrors,
		AccountsCreated: accountsCreated,
		TransfersTotal:  transfersTotal,
		TransferAmount:  transferAmount,
		BalanceChanges:  balanceChanges,
	}, nil
}

func NewStructuredLogger(config Config) *slog.Logger {
	level := slog.LevelInfo
	switch config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})).With(
		"service", config.OtelServiceName,
		"version", config.OtelServiceVersion,
		"environment", config.OtelEnvironment,
	)
}
