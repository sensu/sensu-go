package eventd

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/eventd")
