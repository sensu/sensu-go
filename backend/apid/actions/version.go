package actions

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	apitools "github.com/sensu/sensu-api-tools"
	"github.com/sensu/sensu-go/version"
)

// VersionController exposes actions which a viewer can perform
type VersionController struct {
	clusterVersion string
}

// NewVersionController returns a new VersionController
func NewVersionController(clusterVersion string) VersionController {
	return VersionController{
		clusterVersion: clusterVersion,
	}
}

// GetVersion returns version information
func (v VersionController) GetVersion(ctx context.Context) *corev2.Version {
	return &corev2.Version{
		SensuBackend: version.Semver(),
		APIGroups:    apitools.APIModuleVersions(),
	}
}
