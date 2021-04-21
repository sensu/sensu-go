package asset

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("asset")
