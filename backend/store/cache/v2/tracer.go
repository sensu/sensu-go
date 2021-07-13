package v2

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/store/cache/v2")
