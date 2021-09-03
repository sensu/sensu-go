package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		SecretsProviderManager *secrets.ProviderManager
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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if err := l.Handle(tt.args.ctx, tt.args.ref, tt.args.event, tt.args.mutatedData); (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLegacyAdapter_pipeHandler(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		LicenseGetter          licensing.Getter
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx           context.Context
		handler       *corev2.Handler
		event         *corev2.Event
		mutatedDataFn func(*corev2.Event) []byte
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantNil      bool
		wantStatus   int
		wantOutputFn func([]byte) string
		wantErr      bool
	}{
		{
			name: "can successfully run a pipe handler",
			fields: fields{
				Executor: &command.ExecutionRequest{},
			},
			args: args{
				ctx: context.Background(),
				handler: func() *corev2.Handler {
					handler := corev2.FakeHandlerCommand("cat")
					handler.Type = "pipe"
					return handler
				}(),
				event: corev2.FixtureEvent("test", "test"),
				mutatedDataFn: func(event *corev2.Event) []byte {
					data, _ := json.Marshal(event)
					return data
				},
			},
			wantNil:    false,
			wantStatus: 0,
			wantOutputFn: func(mutatedData []byte) string {
				return string(mutatedData)
			},
			wantErr: false,
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
			mutatedData := tt.args.mutatedDataFn(tt.args.event)

			got, err := l.pipeHandler(tt.args.ctx, tt.args.handler, tt.args.event, mutatedData)
			if (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.pipeHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("LegacyAdapter.pipeHandler() got = %v, wantNil %v", got, tt.wantNil)
				return
			}
			if got != nil {
				if got.Status != tt.wantStatus {
					t.Errorf("LegacyAdapter.pipeHandler() status = %v, want %v", got.Status, tt.wantStatus)
					return
				}
				wantOutput := tt.wantOutputFn(mutatedData)
				if got.Output != wantOutput {
					t.Errorf("LegacyAdapter.pipeHandler() output = %v, want %v", got.Output, wantOutput)
					return
				}
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
