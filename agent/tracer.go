package agent

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("agent")
