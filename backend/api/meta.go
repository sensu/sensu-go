package api

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
)

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) *corev2.Claims {
	if value := ctx.Value(corev2.ClaimsKey); value != nil {
		claims, ok := value.(*corev2.Claims)
		if !ok {
			return nil
		}
		return claims
	}
	return nil
}

// Fills given resource's created-by field using given authorization details.
func setCreatedBy(ctx context.Context, resource corev3.Resource) {
	meta := resource.GetMetadata()
	if claims := GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		resource.SetMetadata(meta)
	}
}
