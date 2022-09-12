package api

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func defaultAuth(store storev2.Interface) *authentication.Authenticator {
	auth := &authentication.Authenticator{}
	provider := &basic.Provider{Store: store, ObjectMeta: corev2.ObjectMeta{Name: basic.Type}}
	auth.AddProvider(provider)
	return auth
}

func defaultStore() storev2.Interface {
	return &mockstore.V2MockStore{}
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
		Store         func() storev2.Interface
		Authenticator func(storev2.Interface) *authentication.Authenticator
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
			Store: func() storev2.Interface {
				user := corev2.FixtureUser("foo")
				store := &mockstore.V2MockStore{}
				store.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
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
			Store: func() storev2.Interface {
				store := &mockstore.V2MockStore{}
				user := corev2.FixtureUser("foo")
				user.PasswordHash, _ = bcrypt.HashPassword("P@ssw0rd!")
				store.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
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
				if test.Error != nil && test.Error != err {
					t.Errorf("bad error: got %v, want %v", err, test.Error)
				}
				t.Fatal("want error, got nil")
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
		Store         func() storev2.Interface
		Authenticator func(storev2.Interface) *authentication.Authenticator
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
			Store: func() storev2.Interface {
				s := &mockstore.V2MockStore{}
				s.On("Get", mock.Anything, mock.Anything).Return(nil, mockError)
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
			Store: func() storev2.Interface {
				s := &mockstore.V2MockStore{}
				user := corev2.FixtureUser("foo")
				user.Password, _ = bcrypt.HashPassword("P@ssw0rd!")
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			},
			Authenticator: defaultAuth,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := test.Store()
			authn := NewAuthenticationClient(test.Authenticator(store))
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
		Store         func() storev2.Interface
		Authenticator func(storev2.Interface) *authentication.Authenticator
		Context       func(*corev2.Claims) context.Context
		WantError     bool
		Error         error
	}{
		{
			Name: "success",
			Store: func() storev2.Interface {
				st := &mockstore.V2MockStore{}
				user := &corev2.User{Username: "foo"}
				st.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
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
