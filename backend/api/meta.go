package api

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// Fills given resource's created-by field using given authorization details.
func setCreatedBy(ctx context.Context, resource corev2.Resource) {
	meta := resource.GetObjectMeta()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		resource.SetObjectMeta(meta)
	}
}
