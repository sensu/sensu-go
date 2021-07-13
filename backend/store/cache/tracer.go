package cache

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/store/cache")
