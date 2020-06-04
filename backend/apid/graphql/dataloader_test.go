package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/graph-gophers/dataloader"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

func contextWithLoadersNoCache(ctx context.Context, cfg ServiceConfig, opts ...dataloader.Option) context.Context {
	opts = append(opts, dataloader.WithCache(&dataloader.NoCache{}))
	return contextWithLoaders(ctx, cfg, opts...)
}

func Test_listAllEvents(t *testing.T) {
	mkEvents := func(num int) []*corev2.Event {
		result := make([]*corev2.Event, num)
		for i := 0; i < num; i++ {
			result[i] = corev2.FixtureEvent("", "")
		}
		return result
	}
	tests := []struct {
		name    string
		setup   func(*MockEventClient)
		wantLen int
		wantErr bool
	}{
		{
			name: "single page",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(500), nil).Once()
			},
			wantLen: 500,
			wantErr: false,
		},
		{
			name: "many pages",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(2000), nil).Once()
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(20), nil).Once()
			},
			wantLen: 2020,
			wantErr: false,
		},
		{
			name: "fetch err",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(2), errors.New("err")).Once()
			},
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockEventClient)
			tt.setup(client)
			got, err := listAllEvents(context.Background(), client)
			if (err != nil) != tt.wantErr {
				t.Errorf("listAllEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("listAllEvents() = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
