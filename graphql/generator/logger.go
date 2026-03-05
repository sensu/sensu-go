package generator

import (
	"github.com/sensu/sensu-go/util/logging"
)

// default logger used in this package
var logger = logging.GetLogger("graphql.generator").WithField("component", "graphql.generator")
