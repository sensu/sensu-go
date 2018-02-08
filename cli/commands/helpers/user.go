package helpers

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
)

func GetCurrentUsername(cfg config.Config) string {
	accessToken := cfg.Tokens().Access
	token, _ := jwt.ParseWithClaims(accessToken, &types.Claims{}, nil)
	claims := token.Claims.(*types.Claims)
	return claims.StandardClaims.Subject
}
