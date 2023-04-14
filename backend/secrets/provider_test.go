package secrets

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	mock.Mock
}

// Get ...
func (m *mockProvider) Get(id string) (string, error) {
	args := m.Called(id)
	return args.Get(0).(string), args.Error(1)
}

// GetObjectMeta ...
func (m *mockProvider) GetMetadata() *corev2.ObjectMeta {
	args := m.Called()
	return args.Get(0).(*corev2.ObjectMeta)
}

// SetObjectMeta ...
func (m *mockProvider) SetMetadata(_ *corev2.ObjectMeta) {}

// RBACName ...
func (m *mockProvider) RBACName() string {
	args := m.Called()
	return args.Get(0).(string)
}

// StorePrefix ...
func (m *mockProvider) StoreName() string {
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

type mockGetter struct {
	mock.Mock
}

// Get ...
func (m *mockGetter) Get(ctx context.Context, name string) (string, string, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

type mockEventReceiver struct {
	mock.Mock
}

func (m *mockEventReceiver) GenerateBackendEvent(component string, status uint32, output string) error {
	args := m.Called(component, status, output)
	return args.Error(0)
}

func TestProviderManager(t *testing.T) {
	mer := &mockEventReceiver{}
	pm := NewProviderManager(mer)
	mp := &mockProvider{}
	mp.On("GetMetadata", mock.Anything).Return(&corev2.ObjectMeta{Name: "env"})
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
	require.Nil(t, pm.Getter)
}
func TestSubSecrets(t *testing.T) {
	ctx := context.Background()

	mg := &mockGetter{}
	mg.On("Get", ctx, "sensu-foo").Return("env", "SENSU_FOO", nil)
	mg.On("Get", ctx, "sensu-baby").Return("vault", "SENSU_BABY", nil)
	mg.On("Get", ctx, "sensu-empty").Return("env", "SENSU_EMPTY", nil)
	mg.On("Get", ctx, "sensu-baz").Return("env", "SENSU_BAZ", nil)
	mg.On("Get", ctx, "sensu-err").Return("", "", ErrSecretNotFound("sensu-err"))
	mg.On("Get", ctx, "sensu-no-provider").Return("foo", "SENSU_NO_PROVIDER", nil)
	mg.On("Get", ctx, "sensu-provider-err").Return("vault", "SENSU_PROVIDER_ERR", nil)

	mer := &mockEventReceiver{}
	mer.On("GenerateBackendEvent", "secrets", uint32(0), msgSecretsProviderOk).Return(nil)
	mer.On("GenerateBackendEvent", "secrets", uint32(0), msgSecretsProviderOk).Return(nil)
	mer.On("GenerateBackendEvent", "secrets", uint32(2), ErrProviderNotAvailable("vault").Error()).Return(nil)

	pm := NewProviderManager(mer)
	pm.Getter = mg

	// no providers with secrets defined returns error
	secretVars, err := pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "FOO",
			Secret: "sensu-foo",
		},
	})
	require.Error(t, err)
	require.Equal(t, 0, len(pm.Providers()))
	require.Equal(t, []string{}, secretVars)

	// create provider env
	env := &mockProvider{}
	env.On("GetMetadata", mock.Anything).Return(&corev2.ObjectMeta{Name: "env"})
	env.On("Get", "SENSU_FOO").Return("bar", nil)
	env.On("Get", "SENSU_EMPTY").Return("", nil)
	env.On("Get", "SENSU_BAZ").Return("boo", nil)
	env.On("Get", "SENSU_ERR").Return("", fmt.Errorf("err on provider"))
	pm.AddProvider(env)
	require.Equal(t, 1, len(pm.Providers()))

	// all found secrets are returned from a single provider
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "FOO",
			Secret: "sensu-foo",
		}, {
			Name:   "EMPTY",
			Secret: "sensu-empty",
		}, {
			Name:   "BAZ",
			Secret: "sensu-baz",
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"FOO=bar", "BAZ=boo"}, secretVars)

	// an error is returned if the provider errors
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "FOO",
			Secret: "sensu-foo",
		}, {
			Name:   "ERR",
			Secret: "sensu-err",
		},
	})
	require.Error(t, err)
	require.Equal(t, []string{}, secretVars)

	// no secrets/no error if no secrets are provided
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{})
	require.NoError(t, err)
	require.Equal(t, []string{}, secretVars)

	// no secrets/no error on nil
	secretVars, err = pm.SubSecrets(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, []string{}, secretVars)

	// create provider vault
	vault := &mockProvider{}
	vault.On("GetMetadata", mock.Anything).Return(&corev2.ObjectMeta{Name: "vault"})
	vault.On("Get", "SENSU_BABY").Return("yoda", nil)
	vault.On("Get", "SENSU_FOO").Return("", nil)
	vault.On("Get", "SENSU_PROVIDER_ERR").Return("", ErrProviderNotAvailable("vault"))
	pm.AddProvider(vault)
	require.Equal(t, 2, len(pm.Providers()))

	// all found secrets are returned from all providers
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "FOO",
			Secret: "sensu-foo",
		}, {
			Name:   "BABY",
			Secret: "sensu-baby",
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"FOO=bar", "BABY=yoda"}, secretVars)

	// provider does not exist
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "NO_PROVIDER",
			Secret: "sensu-no-provider",
		},
	})
	require.Error(t, err)
	require.Equal(t, []string{}, secretVars)

	// provider error getting secret
	secretVars, err = pm.SubSecrets(ctx, []*corev2.Secret{
		{
			Name:   "PROVIDER_ERR",
			Secret: "sensu-provider-err",
		},
	})
	require.Error(t, err)
	require.Equal(t, []string{}, secretVars)
}
