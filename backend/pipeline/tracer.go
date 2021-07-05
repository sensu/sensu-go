package pipeline

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/pipeline")
