package api

import (
	"path"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "backend/api",
})

func logWithAttrs(attrs *authorization.Attributes) *logrus.Entry {
	fields := logrus.Fields{
		"api":           path.Join(attrs.APIGroup, attrs.APIVersion),
		"namespace":     attrs.Namespace,
		"resource_type": attrs.Resource,
		"verb":          attrs.Verb,
		"resource_name": attrs.ResourceName,
		"user":          attrs.User.Username,
	}
	return logger.WithFields(fields)
}
