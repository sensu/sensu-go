package etcd

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("backend/store/etcd")
