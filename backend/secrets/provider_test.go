package secrets

import (
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	mock.Mock
}

// Get ...
func (m *mockProvider) Get(name string) (secret string, err error) {
	args := m.Called(name)
	return args.Get(0).(string), args.Error(1)
}

// GetObjectMeta ...
func (m *mockProvider) GetObjectMeta() corev2.ObjectMeta {
	args := m.Called()
	return args.Get(0).(corev2.ObjectMeta)
}

// SetObjectMeta ...
func (m *mockProvider) SetObjectMeta(meta corev2.ObjectMeta) {}

// RBACName ...
func (m *mockProvider) RBACName() string {
	args := m.Called()
	return args.Get(0).(string)
}

// SetNamespace ...
func (m *mockProvider) SetNamespace(namespace string) {}

// StorePrefix ...
func (m *mockProvider) StorePrefix() string {
	args := m.Called()
	return args.Get(0).(string)
}

// URIPath ...
func (m *mockProvider) URIPath() string {
	args := m.Called()
	return args.Get(0).(string)
}

// Validate ...
func (m *mockProvider) Validate() error {
	args := m.Called()
	return args.Get(0).(error)
}

func TestProviderManager(t *testing.T) {
	pm := NewProviderManager()
	mp := &mockProvider{}
	mp.On("GetObjectMeta", mock.Anything).Return(corev2.ObjectMeta{Name: "env"})
	pm.AddProvider(mp)
	require.NotEmpty(t, pm.Providers())
	require.Equal(t, 1, len(pm.Providers()))
	err := pm.RemoveProvider("env")
	require.NoError(t, err)
	require.Empty(t, pm.Providers())
	require.Equal(t, 0, len(pm.Providers()))
	err = pm.RemoveProvider("env")
	require.Error(t, err)
	require.Empty(t, pm.Providers())
}

func TestSubSecrets(t *testing.T) {
	pm := NewProviderManager()

	// create provider env
	mp1 := &mockProvider{}
	mp1.On("GetObjectMeta", mock.Anything).Return(corev2.ObjectMeta{Name: "env"})
	mp1.On("Get", "foo").Return("bar", nil)
	mp1.On("Get", "baby").Return("", nil)
	mp1.On("Get", "baz").Return("boo", nil)
	mp1.On("Get", "err").Return("", fmt.Errorf("err on provider"))
	pm.AddProvider(mp1)
	require.Equal(t, 1, len(pm.Providers()))

	// all found secrets are returned from a single provider
	secretVars, err := pm.SubSecrets([]*corev2.Secret{
		&corev2.Secret{
			Name:   "FOO",
			Secret: "foo",
		},
		&corev2.Secret{
			Name:   "BABY",
			Secret: "baby",
		},
		&corev2.Secret{
			Name:   "BAZ",
			Secret: "baz",
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"FOO=bar", "BAZ=boo"}, secretVars)

	// an error is returned if the provider errors
	secretVars, err = pm.SubSecrets([]*corev2.Secret{
		&corev2.Secret{
			Name:   "FOO",
			Secret: "foo",
		},
		&corev2.Secret{
			Name:   "ERR",
			Secret: "err",
		},
	})
	require.Error(t, err)
	require.Equal(t, []string{}, secretVars)

	// no secrets/no error if no secrets are provided
	secretVars, err = pm.SubSecrets([]*corev2.Secret{})
	require.NoError(t, err)
	require.Equal(t, []string{}, secretVars)

	// no secrets/no error on nil
	secretVars, err = pm.SubSecrets(nil)
	require.NoError(t, err)
	require.Equal(t, []string{}, secretVars)

	// create provider vault
	mp2 := &mockProvider{}
	mp2.On("GetObjectMeta", mock.Anything).Return(corev2.ObjectMeta{Name: "vault"})
	mp2.On("Get", "baby").Return("yoda", nil)
	mp2.On("Get", "foo").Return("", nil)
	pm.AddProvider(mp2)
	require.Equal(t, 2, len(pm.Providers()))

	// all found secrets are returned from all providers
	secretVars, err = pm.SubSecrets([]*corev2.Secret{
		&corev2.Secret{
			Name:   "FOO",
			Secret: "foo",
		},
		&corev2.Secret{
			Name:   "BABY",
			Secret: "baby",
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"FOO=bar", "BABY=yoda"}, secretVars)

	// no secrets/no error with no providers
	require.NoError(t, pm.RemoveProvider("env"))
	require.NoError(t, pm.RemoveProvider("vault"))
	secretVars, err = pm.SubSecrets([]*corev2.Secret{
		&corev2.Secret{
			Name:   "FOO",
			Secret: "foo",
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{}, secretVars)
}
