package etcd

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

var grpcLogger = logrus.WithField("component", "grpc")

// setGRPCLogger sets the logger for the entire gRPC library to use logrus
func setGRPCLogger() {
	v2Logger := grpclog.NewLoggerV2(
		ioutil.Discard,
		grpcLogger.WriterLevel(logrus.WarnLevel),
		grpcLogger.WriterLevel(logrus.ErrorLevel))
	grpclog.SetLoggerV2(v2Logger)
}
