package basic

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestProviderAuthenticate(t *testing.T) {
	tests := []struct {
		Name     string
		Store    storev2.Interface
		Username string
		Password string
		Hash     string
		Error    bool
	}{
		{
			Name:     "username not supplied",
			Username: "",
			Password: "hello",
			Error:    true,
		},
		{
			Name:     "password not supplied",
			Username: "Hello",
			Password: "",
			Error:    true,
		},
		{
			Name: "user not found",
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				s.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})
				return s
			}(),
			Username: "eric",
			Password: "password",
			Error:    true,
		},
		{
			Name: "user disabled",
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				user := &corev2.User{
					Username: "eric",
					Disabled: true,
				}
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			}(),
			Username: "eric",
			Password: "password",
			Error:    true,
		},
		{
			Name: "password hash does not match",
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				pw, err := bcrypt.HashPassword("password")
				if err != nil {
					panic(err)
				}
				user := &corev2.User{
					Username:     "eric",
					PasswordHash: pw,
				}
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			}(),
			Username: "eric",
			Password: "asldjflkdsajfl",
			Error:    true,
		},
		{
			Name: "password hash does match",
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				pw, err := bcrypt.HashPassword("password")
				if err != nil {
					panic(err)
				}
				user := &corev2.User{
					Username:     "eric",
					PasswordHash: pw,
				}
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			}(),
			Username: "eric",
			Password: "password",
			Error:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			provider := &Provider{
				Store: test.Store,
			}
			claims, err := provider.Authenticate(context.Background(), test.Username, test.Password)
			if err != nil {
				if !test.Error {
					t.Fatal(err)
				}
			} else {
				if test.Error {
					t.Fatal("want non-nil error")
				}
				if claims == nil {
					t.Fatal("nil claims")
				}
			}
		})
	}
}

func TestProviderRefresh(t *testing.T) {
	tests := []struct {
		Name   string
		Claims *corev2.Claims
		Store  storev2.Interface
		Error  bool
	}{
		{
			Name: "user not found",
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				s.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})
				return s
			}(),
			Claims: &corev2.Claims{
				Provider: corev2.AuthProviderClaims{
					UserID: "eric",
				},
			},
			Error: true,
		},
		{
			Name: "user disabled",
			Claims: &corev2.Claims{
				Provider: corev2.AuthProviderClaims{
					UserID: "eric",
				},
			},
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				user := &corev2.User{
					Username: "eric",
					Disabled: true,
				}
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			}(),
			Error: true,
		},
		{
			Name: "happy",
			Claims: &corev2.Claims{
				Provider: corev2.AuthProviderClaims{
					UserID: "eric",
				},
			},
			Store: func() storev2.Interface {
				s := new(mockstore.V2MockStore)
				pw, err := bcrypt.HashPassword("password")
				if err != nil {
					panic(err)
				}
				user := &corev2.User{
					Username:     "eric",
					PasswordHash: pw,
				}
				s.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)
				return s
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			provider := &Provider{
				Store: test.Store,
			}
			claims, err := provider.Refresh(context.Background(), test.Claims)
			if err != nil {
				if !test.Error {
					t.Fatal(err)
				}
			} else {
				if test.Error {
					t.Fatal("want non-nil error")
				}
				if claims == nil {
					t.Fatal("nil claims")
				}
			}
		})
	}
}
