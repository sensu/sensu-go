package ringv2

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/ringv2")
