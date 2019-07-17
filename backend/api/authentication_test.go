package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
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
				store.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(nil)
				store.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, nil)
				return store
			},
			Authenticator: defaultAuth,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(store, test.Authenticator(store))
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
				store.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(nil)
				store.On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").Return(user, nil)
				return store
			},
			Authenticator: defaultAuth,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(store, test.Authenticator(store))
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

func TestLogout(t *testing.T) {
	tests := []struct {
		Name          string
		Store         func() store.Store
		Authenticator func(store.Store) *authentication.Authenticator
		Context       func(*corev2.Claims) context.Context
		WantError     bool
		Error         error
	}{
		{
			Name:          "invalid token",
			Store:         defaultStore,
			Authenticator: defaultAuth,
			Context:       func(*corev2.Claims) context.Context { return context.Background() },
			WantError:     true,
			Error:         corev2.ErrInvalidToken,
		},
		{
			Name: "not whitelisted",
			Store: func() store.Store {
				store := &mockstore.MockStore{}
				store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(fmt.Errorf("error"))
				return store
			},
			Authenticator: defaultAuth,
			Context:       contextWithClaims,
			WantError:     true,
		},
		{
			Name: "success",
			Store: func() store.Store {
				store := &mockstore.MockStore{}
				store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)
				return store
			},
			Authenticator: defaultAuth,
			Context:       contextWithClaims,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			claims := corev2.FixtureClaims("foo", nil)
			ctx := test.Context(claims)
			store := test.Store()
			authenticator := test.Authenticator(store)
			auth := NewAuthenticationClient(store, authenticator)
			err := auth.Logout(ctx)
			if err == nil && test.WantError {
				t.Fatal("got non-nil error")
			}
			if err != nil && !test.WantError {
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
			Name: "not whitelisted",
			Store: func() store.Store {
				st := &mockstore.MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On(
					"GetToken",
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&corev2.Claims{}, &store.ErrNotFound{})
				st.On("GetUser",
					mock.Anything,
					mock.AnythingOfType("string")).Return(user, nil)
				return st
			},
			Authenticator: defaultAuth,
			Context:       contextWithClaims,
			WantError:     true,
			Error:         corev2.ErrUnauthorized,
		},
		{
			Name: "cannot whitelist access token",
			Store: func() store.Store {
				st := &mockstore.MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(fmt.Errorf("error"))
				st.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)
				st.On("GetToken",
					mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&corev2.Claims{}, nil)
				st.On("GetUser",
					mock.Anything, mock.AnythingOfType("string")).Return(user, nil)
				return st
			},
			Authenticator: defaultAuth,
			Context:       contextWithClaims,
			WantError:     true,
		},
		{
			Name: "success",
			Store: func() store.Store {
				st := &mockstore.MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(nil)
				st.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)
				st.On("GetToken",
					mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(&types.Claims{}, nil)
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
			auth := NewAuthenticationClient(store, authenticator)
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
