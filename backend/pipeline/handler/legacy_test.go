package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockassetgetter"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/testing/mocklicense"
	"github.com/sensu/sensu-go/testing/mocksecrets"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestHelperHandlerProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_HANDLER_PROCESS") != "1" {
		return
	}

	command := strings.Join(os.Args[3:], " ")
	stdin, _ := ioutil.ReadAll(os.Stdin)

	switch command {
	case "cat":
		fmt.Fprintf(os.Stdout, "%s", stdin)
	}
	os.Exit(0)
}

func TestLegacyAdapter_Name(t *testing.T) {
	o := &LegacyAdapter{}
	want := "LegacyAdapter"

	if got := o.Name(); want != got {
		t.Errorf("LegacyAdapter.Name() = %v, want %v", got, want)
	}
}

func TestLegacyAdapter_CanHandle(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		LicenseGetter          licensing.Getter
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "returns false when resource reference is not a core/v2.Handler",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.Handler",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &LegacyAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				LicenseGetter:          tt.fields.LicenseGetter,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			if got := h.CanHandle(tt.args.ref); got != tt.want {
				t.Errorf("LegacyAdapter.CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacyAdapter_Handle(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		LicenseGetter          licensing.Getter
		SecretsProviderManager secrets.ProviderManagerer
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx         context.Context
		ref         *corev2.ResourceReference
		event       *corev2.Event
		mutatedData []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error if a handler cannot be fetched from the store",
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					Name: "handler1",
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
			},
			fields: fields{
				Store: func() store.Store {
					var handler *corev2.Handler
					err := errors.New("not found")
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, err)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "failed to fetch handler from store: not found",
		},
		{
			name: "returns an error if a pipe handler returns an error",
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					Name: "handler1",
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
			},
			fields: fields{
				SecretsProviderManager: func() secrets.ProviderManagerer {
					var secrets []string
					manager := &mocksecrets.ProviderManager{}
					manager.On("SubSecrets", mock.Anything, mock.Anything).Return(secrets, errors.New("secrets error"))
					return manager
				}(),
				Store: func() store.Store {
					handler := corev2.FixtureHandler("handler1")
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "secrets error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LegacyAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				LicenseGetter:          tt.fields.LicenseGetter,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			err := l.Handle(tt.args.ctx, tt.args.ref, tt.args.event, tt.args.mutatedData)
			if (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("LegacyAdapter.Handle() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestLegacyAdapter_pipeHandler(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		LicenseGetter          licensing.Getter
		SecretsProviderManager secrets.ProviderManagerer
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx         context.Context
		handler     *corev2.Handler
		event       *corev2.Event
		mutatedData []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *command.ExecutionResponse
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "appends the license file to envvars when a licensegetter and license exist",
			fields: fields{
				LicenseGetter: func() licensing.Getter {
					getter := &mocklicense.Getter{}
					getter.On("Get").Return("foo-license")
					return getter
				}(),
				Executor: func() command.Executor {
					ex := &mockexecutor.MockExecutor{}
					ex.SetRequestFunc(func(_ context.Context, request command.ExecutionRequest) {
						for _, env := range request.Env {
							if env == "SENSU_LICENSE_FILE=foo-license" {
								response := command.FixtureExecutionResponse(0, "")
								ex.UnsafeReturn(response, nil)
								return
							}
						}
						response := command.FixtureExecutionResponse(1, "")
						ex.UnsafeReturn(response, nil)
					})
					return ex
				}(),
			},
			args: args{
				ctx:         context.Background(),
				handler:     corev2.FixtureHandler("handler1"),
				event:       corev2.FixtureEvent("entity1", "check1"),
				mutatedData: []byte{},
			},
			want: command.FixtureExecutionResponse(0, ""),
		},
		{
			name: "returns an error if secret retrieval fails",
			fields: fields{
				SecretsProviderManager: func() secrets.ProviderManagerer {
					manager := &mocksecrets.ProviderManager{}
					manager.On("SubSecrets", mock.Anything, mock.Anything).Return([]string{}, errors.New("secrets error"))
					return manager
				}(),
			},
			args: args{
				ctx:         context.Background(),
				handler:     corev2.FixtureHandler("handler1"),
				event:       corev2.FixtureEvent("entity1", "check1"),
				mutatedData: []byte{},
			},
			wantErr:    true,
			wantErrMsg: "secrets error",
		},
		{
			name: "secrets are added to the execution environment",
			fields: fields{
				SecretsProviderManager: func() secrets.ProviderManagerer {
					manager := &mocksecrets.ProviderManager{}
					manager.On("SubSecrets",
						mock.Anything,
						mock.MatchedBy(func(secrets []*corev2.Secret) bool {
							for _, secret := range secrets {
								if secret.Name == "MYSECRET" && secret.Secret == "topsecret" {
									return true
								}
							}
							return false
						})).
						Return([]string{"MYSECRET=topsecret"}, nil)
					return manager
				}(),
				Executor: func() command.Executor {
					ex := &mockexecutor.MockExecutor{}
					ex.SetRequestFunc(func(_ context.Context, request command.ExecutionRequest) {
						for _, env := range request.Env {
							if env == "MYSECRET=topsecret" {
								response := command.FixtureExecutionResponse(0, "")
								ex.UnsafeReturn(response, nil)
								return
							}
						}
						response := command.FixtureExecutionResponse(1, "")
						ex.UnsafeReturn(response, nil)
					})
					return ex
				}(),
			},
			args: args{
				ctx: context.Background(),
				handler: func() *corev2.Handler {
					handler := corev2.FixtureHandler("handler1")
					handler.Secrets = []*corev2.Secret{
						{
							Name:   "MYSECRET",
							Secret: "topsecret",
						},
					}
					return handler
				}(),
				event:       corev2.FixtureEvent("entity1", "check1"),
				mutatedData: []byte{},
			},
			want: command.FixtureExecutionResponse(0, ""),
		},
		// TODO: add a test here for when asset.GetAssets() returns errors. The
		// asset.GetAssets() function does not currently return errors and
		// only logs them.
		// See issue #4407: https://github.com/sensu/sensu-go/issues/4407
		{
			name: "asset environment variables are added to the execution environment",
			fields: fields{
				AssetGetter: func() asset.Getter {
					getter := &mockassetgetter.MockAssetGetter{}
					runtimeAsset := &asset.RuntimeAsset{
						Name: "asset1",
						Path: "/path/to/asset1",
					}
					getter.On("Get",
						mock.Anything,
						mock.MatchedBy(func(asset *corev2.Asset) bool {
							return asset.Name == "asset1"
						})).
						Return(runtimeAsset, nil)
					return getter
				}(),
				Executor: func() command.Executor {
					ex := &mockexecutor.MockExecutor{}
					ex.SetRequestFunc(func(_ context.Context, request command.ExecutionRequest) {
						for _, env := range request.Env {
							if env == "ASSET1_PATH=/path/to/asset1" {
								response := command.FixtureExecutionResponse(0, "")
								ex.UnsafeReturn(response, nil)
								return
							}
						}
						response := command.FixtureExecutionResponse(1, "")
						ex.UnsafeReturn(response, nil)
					})
					return ex
				}(),
				Store: func() store.Store {
					asset := corev2.FixtureAsset("asset1")
					stor := &mockstore.MockStore{}
					stor.On("GetAssetByName", mock.Anything, "asset1").
						Return(asset, nil)
					return stor
				}(),
			},
			args: args{
				ctx: context.Background(),
				handler: func() *corev2.Handler {
					handler := corev2.FixtureHandler("handler1")
					handler.RuntimeAssets = []string{"asset1"}
					return handler
				}(),
				event:       corev2.FixtureEvent("entity1", "check1"),
				mutatedData: []byte{},
			},
			want: command.FixtureExecutionResponse(0, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LegacyAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				LicenseGetter:          tt.fields.LicenseGetter,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			got, err := l.pipeHandler(tt.args.ctx, tt.args.handler, tt.args.event, tt.args.mutatedData)
			if (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.pipeHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("LegacyAdapter.pipeHandler() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LegacyAdapter.pipeHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacyAdapter_socketHandlerTCP(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	ctx := context.Background()
	event := corev2.FixtureEvent("test", "test")
	mutatedData, _ := json.Marshal(event)
	handler := &corev2.Handler{
		Type: "tcp",
		Socket: &corev2.HandlerSocket{
			Host: "127.0.0.1",
			Port: 5678,
		},
	}

	l := &LegacyAdapter{}

	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer func() {
			require.NoError(t, listener.Close())
		}()

		ready <- struct{}{}

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() {
			require.NoError(t, conn.Close())
		}()

		buffer, err := ioutil.ReadAll(conn)
		if err != nil {
			return
		}

		assert.Equal(t, mutatedData, buffer)
		done <- struct{}{}
	}()

	<-ready
	_, err := l.socketHandler(ctx, handler, event, mutatedData)

	assert.NoError(t, err)
	<-done
}

func TestLegacyAdapter_socketHandlerUDP(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	ctx := context.Background()
	event := corev2.FixtureEvent("test", "test")
	mutatedData, _ := json.Marshal(event)
	handler := &corev2.Handler{
		Type: "udp",
		Socket: &corev2.HandlerSocket{
			Host: "127.0.0.1",
			Port: 5678,
		},
	}

	l := &LegacyAdapter{}

	go func() {
		listener, err := net.ListenPacket("udp", ":5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer func() {
			require.NoError(t, listener.Close())
		}()

		ready <- struct{}{}

		buffer := make([]byte, 8192)
		rlen, _, err := listener.ReadFrom(buffer)

		assert.NoError(t, err)
		assert.Equal(t, mutatedData, buffer[0:rlen])
		done <- struct{}{}
	}()

	<-ready

	_, err := l.socketHandler(ctx, handler, event, mutatedData)

	assert.NoError(t, err)
	<-done
}
