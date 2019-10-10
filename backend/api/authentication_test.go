package api

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	refreshClaims := &corev2.Claims{StandardClaims: corev2.StandardClaims(claims.Subject)}
	ctx := context.Background()
	ctx = context.WithValue(ctx, corev2.AccessTokenClaims, claims)
	ctx = context.WithValue(ctx, corev2.RefreshTokenClaims, refreshClaims)
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
				return store
			},
			Authenticator: defaultAuth,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(test.Authenticator(store))
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
				return store
			},
			Authenticator: defaultAuth,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(test.Authenticator(store))
			err := authn.TestCreds(test.Context(), test.Username, test.Password)
			if test.WantError && err == nil {
				t.Fatal("want error, got nil")
				if test.Error != nil && test.Error != err {
					t.Fatalf("bad error: got %v, want %v", err, test.Error)
				}
			}
			if !test.WantError && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRefreshAccessToken(t *testing.T) {
	tests := []struct {
		Name          string
		Store         func() store.Store
		Authenticator func(store.Store) *authentication.Authenticator
		Context       func(*corev2.Claims) context.Context
		WantError     bool
		Error         error
	}{
		{
			Name: "success",
			Store: func() store.Store {
				st := &mockstore.MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On("GetUser",
					mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"),
				).Return(user, nil)
				return st
			},
			Authenticator: defaultAuth,
			Context: func(claims *corev2.Claims) context.Context {
				ctx := contextWithClaims(claims)
				_, refreshTokenString, _ := jwt.RefreshToken(ctx.Value(corev2.RefreshTokenClaims).(*corev2.Claims))
				ctx = context.WithValue(ctx, corev2.RefreshTokenString, refreshTokenString)
				return ctx
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			claims := corev2.FixtureClaims("foo", nil)
			ctx := test.Context(claims)
			store := test.Store()
			authenticator := test.Authenticator(store)
			auth := NewAuthenticationClient(authenticator)
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
