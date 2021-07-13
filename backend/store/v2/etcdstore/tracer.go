package etcdstore

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/store/v2/etcdstore")
