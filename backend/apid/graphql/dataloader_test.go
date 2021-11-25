package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/graph-gophers/dataloader"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/mock"
)

func contextWithLoadersNoCache(ctx context.Context, cfg ServiceConfig, opts ...dataloader.Option) context.Context {
	opts = append(opts, dataloader.WithCache(&dataloader.NoCache{}))
	return contextWithLoaders(ctx, cfg, opts...)
}

func Test_listEvents(t *testing.T) {
	mkEvents := func(num int) []*corev2.Event {
		result := make([]*corev2.Event, num)
		for i := 0; i < num; i++ {
			result[i] = corev2.FixtureEvent("", "")
		}
		return result
	}
	tests := []struct {
		name    string
		entity  string
		maxSize int
		setup   func(*MockEventClient)
		wantLen int
		wantErr bool
	}{
		{
			name: "single page",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(500), nil).Once()
			},
			maxSize: 5_000,
			wantLen: 500,
			wantErr: false,
		},
		{
			name: "many pages",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(2000), nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*store.SelectionPredicate)
					arg.Continue = "test"
				}).Once()
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(20), nil).Once()
			},
			maxSize: 5_000,
			wantLen: 2020,
			wantErr: false,
		},
		{
			name:   "single page",
			entity: "wiggum",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Panic("called wrong method")
				c.On("ListEventsByEntity", mock.Anything, "wiggum", mock.Anything).Return(mkEvents(500), nil).Once()
			},
			maxSize: 5_000,
			wantLen: 500,
			wantErr: false,
		},
		{
			name: "fetch err",
			setup: func(c *MockEventClient) {
				c.On("ListEvents", mock.Anything, mock.Anything).Return(mkEvents(2), errors.New("err")).Once()
			},
			maxSize: 5_000,
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockEventClient)
			tt.setup(client)
			got, err := listEvents(context.Background(), client, tt.entity, tt.maxSize)
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

func Test_listEntities(t *testing.T) {
	mkEntities := func(num int) []*corev2.Entity {
		result := make([]*corev2.Entity, num)
		for i := 0; i < num; i++ {
			result[i] = corev2.FixtureEntity("asdf")
		}
		return result
	}
	tests := []struct {
		name    string
		maxLen  int
		setup   func(*MockEntityClient)
		wantLen int
		wantErr bool
	}{
		{
			name: "single page",
			setup: func(c *MockEntityClient) {
				c.On("ListEntities", mock.Anything, mock.Anything).Return(mkEntities(500), nil).Once()
			},
			maxLen:  10_000,
			wantLen: 500,
			wantErr: false,
		},
		{
			name: "many pages",
			setup: func(c *MockEntityClient) {
				c.On("ListEntities", mock.Anything, mock.Anything).Return(mkEntities(2000), nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*store.SelectionPredicate)
					arg.Continue = "test"
				}).Once()
				c.On("ListEntities", mock.Anything, mock.Anything).Return(mkEntities(20), nil).Once()
			},
			maxLen:  10_000,
			wantLen: 2020,
			wantErr: false,
		},
		{
			name: "hit upper bounds",
			setup: func(c *MockEntityClient) {
				c.On("ListEntities", mock.Anything, mock.Anything).Return(mkEntities(1000), nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*store.SelectionPredicate)
					arg.Continue = "test"
				})
			},
			maxLen:  2500,
			wantLen: 3000,
			wantErr: false,
		},
		{
			name: "fetch err",
			setup: func(c *MockEntityClient) {
				c.On("ListEntities", mock.Anything, mock.Anything).Return(mkEntities(2), errors.New("err")).Once()
			},
			maxLen:  10_000,
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockEntityClient)
			tt.setup(client)
			got, err := listEntities(context.Background(), client, tt.maxLen)
			if (err != nil) != tt.wantErr {
				t.Errorf("listEntities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("listEntities() = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
