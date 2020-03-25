package helpers

import (
	jwt "github.com/dgrijalva/jwt-go"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli/client/config"
)

// GetCurrentUsername retrieves the username from the active JWT
func GetCurrentUsername(cfg config.Config) string {
	tokens := cfg.Tokens()
	if tokens == nil {
		return ""
	}

	accessToken := tokens.Access
	token, _ := jwt.ParseWithClaims(accessToken, &corev2.Claims{}, nil)
	if token == nil {
		return ""
	}
	claims := token.Claims.(*corev2.Claims)
	return claims.StandardClaims.Subject
}
