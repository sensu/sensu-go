package v2_test

import (
	"context"
	"testing"

	corev3 "github.com/sensu/core/v3"
	apitools "github.com/sensu/sensu-api-tools"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestWatcherOf(t *testing.T) {
	sto := new(mockstore.V2MockStore)
	cs := new(mockstore.ConfigStore)
	sto.On("GetConfigStore").Return(cs)
	type myInterface interface {
		corev3.Resource
		GetCommand() string
	}
	ch := make(chan []storev2.WatchEvent)
	cs.On("Watch", mock.Anything, mock.Anything).Return((<-chan []storev2.WatchEvent)(ch))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := storev2.WatcherOf[myInterface](ctx, sto)
	numberOfWatchers := len(apitools.FindTypesOf[myInterface]())
	cs.AssertNumberOfCalls(t, "Watch", numberOfWatchers)
	go func() {
		ch <- nil
		ch <- nil
	}()
	<-watcher.Result()
	<-watcher.Result()
}
