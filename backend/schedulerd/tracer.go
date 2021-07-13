package schedulerd

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/schedulerd")
