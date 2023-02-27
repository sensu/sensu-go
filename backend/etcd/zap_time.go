package etcd

import zapcore "go.uber.org/zap/zapcore"

const precisionTimeLayout = "2006-01-02T15:04:05.999Z07:00"

// PrecisionTimeEncoder serializes a time.Time to an RFC3339-formatted string
// with millisecond precision.
var PrecisionTimeEncoder = zapcore.TimeEncoderOfLayout(precisionTimeLayout)
