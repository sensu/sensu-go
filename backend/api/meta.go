package api

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// Fills given resource's created-by field using given authorization details.
func setCreatedBy(ctx context.Context, resource corev3.Resource) {
	meta := resource.GetMetadata()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		resource.SetMetadata(meta)
	}
}
