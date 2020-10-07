package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// Fills given resource's created-by field using given authorization details.
func setCreatedBy(ctx context.Context, res corev2.Resource) {
	meta := res.GetObjectMeta()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		res.SetObjectMeta(meta)
	}
}
