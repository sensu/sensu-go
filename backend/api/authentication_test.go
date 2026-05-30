package api

import (
	"context"
	"errors"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func defaultAuth(store store.Store) *authentication.Authenticator {
	auth := &authentication.Authenticator{}
	provider := &basic.Provider{Store: store, ObjectMeta: corev2.ObjectMeta{Name: basic.Type}}
	auth.AddProvider(provider)
	return auth
}

func defaultStore() store.Store {
	return &mockstore.MockStore{}
}

func contextWithClaims(claims *corev2.Claims) context.Context {
	refreshClaims := &corev2.Claims{
		StandardClaims: corev2.StandardClaims(claims.Subject),
		SessionID:      claims.SessionID,
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, corev2.AccessTokenClaims, claims)
	ctx = context.WithValue(ctx, corev2.RefreshTokenClaims, refreshClaims)

	ctx = context.WithValue(ctx, "accessTokenExpiry", 5*time.Minute)
	ctx = context.WithValue(ctx, "refreshTokenExpiry", 12*time.Hour)

	return ctx
}

func TestCreateAccessToken(t *testing.T) {
	tests := []struct {
		Name          string
		Username      string
		Password      string
		Store         func() store.Store
		Authenticator func(store.Store) *authentication.Authenticator
		Context       func() context.Context
		WantError     bool
		Error         error
	}{
		{
			Name:          "no credentials",
			Store:         defaultStore,
			Authenticator: defaultAuth,
			Context:       context.Background,
			WantError:     true,
			Error:         corev2.ErrUnauthorized,
		},
		{
			Name:     "invalid credentials",
			Username: "foo",
			Password: "P@ssw0rd!",
			Store: func() store.Store {
				user := corev2.FixtureUser("foo")
				store := &mockstore.MockStore{}
				store.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, errors.New("error"))
				store.On("UpdateSession", mock.Anything, "foo", mock.Anything, mock.Anything).Return(nil)
				return store
			},
			Authenticator: defaultAuth,
			Context:       context.Background,
			WantError:     true,
			Error:         corev2.ErrUnauthorized,
		},
		{
			Name:     "success",
			Username: "foo",
			Password: "P@ssw0rd!",
			Context:  context.Background,
			Store: func() store.Store {
				store := &mockstore.MockStore{}
				user := corev2.FixtureUser("foo")
				store.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, nil)
				store.On("UpdateSession", mock.Anything, "foo", mock.Anything, mock.Anything).Return(nil)
				return store
			},
			Authenticator: defaultAuth,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(test.Authenticator(store), store)
			tokens, err := authn.CreateAccessToken(test.Context(), test.Username, test.Password)
			if test.WantError && err == nil {
				t.Fatal("want error, got nil")
				if test.Error != nil && test.Error != err {
					t.Fatalf("bad error: got %v, want %v", err, test.Error)
				}
			}
			if !test.WantError && err != nil {
				t.Fatal(err)
			}
			if tokens != nil {
				if err := tokens.Validate(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestTestCreds(t *testing.T) {
	mockError := errors.New("error")

	tests := []struct {
		Name          string
		Username      string
		Password      string
		Store         func() store.Store
		Authenticator func(store.Store) *authentication.Authenticator
		Context       func() context.Context
		WantError     bool
		Error         error
	}{
		{
			Name:          "no credentials",
			Store:         defaultStore,
			Authenticator: defaultAuth,
			Context:       context.Background,
			WantError:     true,
			Error:         basic.ErrEmptyUsernamePassword,
		},
		{
			Name:     "bubble up authentication error",
			Username: "foo",
			Password: "P@ssw0rd!",
			Store: func() store.Store {
				s := &mockstore.MockStore{}
				user := corev2.FixtureUser("foo")
				s.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, mockError)
				return s
			},
			Authenticator: defaultAuth,
			Context:       context.Background,
			WantError:     true,
			Error:         mockError,
		},
		{
			Name:     "success",
			Username: "foo",
			Password: "P@ssw0rd!",
			Context:  context.Background,
			Store: func() store.Store {
				s := &mockstore.MockStore{}
				user := corev2.FixtureUser("foo")
				s.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, nil)
				return s
			},
			Authenticator: defaultAuth,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(test.Authenticator(store), store)
			err := authn.TestCreds(test.Context(), test.Username, test.Password)

			if test.WantError && test.Error != err {
				t.Fatalf("bad error: got %v, want %v", err, test.Error)
			}

			if !test.WantError && err != nil {
				t.Fatalf("expected no error, go %v", err)
			}
		})
	}
}

func TestRefreshAccessToken(t *testing.T) {
	tests := []struct {
		Name          string
		Store         func(string) store.Store
		Authenticator func(store.Store) *authentication.Authenticator
		Context       func(*corev2.Claims) (context.Context, string)
		WantError     bool
		Error         error
	}{
		{
			Name: "success",
			Store: func(refreshTokenId string) store.Store {
				st := &mockstore.MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On("GetUser",
					mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"),
				).Return(user, nil)
				st.On("GetSession",
					mock.AnythingOfType("*context.valueCtx"), user.Username, mock.AnythingOfType("string"),
				).Return(refreshTokenId, nil)
				st.On("UpdateSession",
					mock.AnythingOfType("*context.valueCtx"), user.Username, mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
				return st
			},
			Authenticator: defaultAuth,
			Context: func(claims *corev2.Claims) (context.Context, string) {
				ctx := contextWithClaims(claims)

				// append configured access token expiry to claims
				var refreshTokenExpiry time.Duration
				if refreshTokenExp := ctx.Value("refreshTokenExpiry"); refreshTokenExp != nil {
					refreshTokenExpiry = refreshTokenExp.(time.Duration)
				}

				refreshToken, refreshTokenString, _ := jwt.RefreshToken(ctx.Value(corev2.RefreshTokenClaims).(*corev2.Claims), jwt.WithRefreshTokenExpiry(refreshTokenExpiry))
				refreshTokenClaims, _ := jwt.GetClaims(refreshToken)
				ctx = context.WithValue(ctx, corev2.RefreshTokenString, refreshTokenString)
				return ctx, refreshTokenClaims.Id
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			claims := corev2.FixtureClaims("foo", nil)
			ctx, refreshTokenId := test.Context(claims)
			store := test.Store(refreshTokenId)
			authenticator := test.Authenticator(store)
			auth := NewAuthenticationClient(authenticator, store)
			_, err := auth.RefreshAccessToken(ctx)
			if err == nil && test.WantError {
				t.Fatal("got non-nil error")
			}
			if err != nil && !test.WantError {
				t.Fatal(err)
			}
		})
	}
}
