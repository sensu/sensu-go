package agentd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/mocktransport"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type failingPublishBus struct {
	*messaging.WizardBus
	subscribers map[string]messaging.Subscriber
}

func (b *failingPublishBus) Subscribe(topic, consumer string, s messaging.Subscriber) (messaging.Subscription, error) {
	b.subscribers[topic] = s
	return b.WizardBus.Subscribe(topic, consumer, s)
}

func (b *failingPublishBus) Publish(topic string, msg interface{}) error {
	if topic == messaging.TopicKeepalive {
		return errors.New("publish failed")
	}
	return b.WizardBus.Publish(topic, msg)
}

func TestNewSessionErrorRemovesUserConfigSubscription(t *testing.T) {
	wizard, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, wizard.Start())
	defer func() { _ = wizard.Stop() }()
	bus := &failingPublishBus{WizardBus: wizard, subscribers: map[string]messaging.Subscriber{}}

	st := &mockstore.MockStore{}
	st.On("GetNamespace", mock.Anything, "acme").Return(&corev2.Namespace{}, nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		User:          "someuser",
		Subscriptions: []string{"testing"},
		Conn:          new(mocktransport.MockTransport),
		Bus:           bus,
		Store:         st,
		Unmarshal:     proto.Unmarshal,
		Marshal:       proto.Marshal,
	}
	session, err := NewSession(context.Background(), cfg)
	require.Error(t, err)
	assert.Nil(t, session)

	topic := messaging.UserConfigTopic("someuser")
	sub, ok := bus.subscribers[topic]
	require.True(t, ok, "user config topic was never subscribed to")

	require.NoError(t, wizard.Publish(topic, "stale"))
	select {
	case <-sub.(*userConfig).subscription:
	case msg := <-sub.(*userConfig).updatesChannel:
		t.Fatalf("stale user-config subscriber still received %v after NewSession failed", msg)
	case <-time.After(200 * time.Millisecond):
	}
}

type failingSubscribeBus struct {
	*messaging.WizardBus
}

func (b *failingSubscribeBus) Subscribe(string, string, messaging.Subscriber) (messaging.Subscription, error) {
	return messaging.Subscription{}, errors.New("subscribe failed")
}

func TestWebSocketHandlerClosesConnWhenNewSessionFails(t *testing.T) {
	wizard, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, wizard.Start())
	defer func() { _ = wizard.Stop() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := &Agentd{
		ctx:            ctx,
		bus:            &failingSubscribeBus{WizardBus: wizard},
		backendEntity:  corev2.FixtureEntity("backend"),
		namespaceCache: cache.NewFromResources([]corev2.Resource{corev2.FixtureNamespace("acme")}, false),
		errChan:        make(chan error, 1),
	}
	ts := httptest.NewServer(http.HandlerFunc(a.webSocketHandler))
	defer ts.Close()

	header := http.Header{}
	header.Set(transport.HeaderKeyAgentName, "testing")
	header.Set(transport.HeaderKeyNamespace, "acme")
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	// A closed server side shows up as a read error; a leaked one blocks until the deadline.
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	_, _, err = conn.ReadMessage()
	require.Error(t, err)
	var netErr net.Error
	if errors.As(err, &netErr) {
		assert.False(t, netErr.Timeout(), "server did not close the hijacked connection after NewSession failed")
	}
}

func TestNewSessionBurialTimeoutRemovesUserConfigSubscription(t *testing.T) {
	old := burialTimeout
	burialTimeout = 100 * time.Millisecond
	defer func() { burialTimeout = old }()

	wizard, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, wizard.Start())
	defer func() { _ = wizard.Stop() }()
	bus := &recordingBus{WizardBus: wizard, subscribers: map[string]messaging.Subscriber{}}

	cfg := SessionConfig{
		AgentName:      "testing",
		Namespace:      "acme",
		User:           "someuser",
		Subscriptions:  []string{"testing"},
		Conn:           new(mocktransport.MockTransport),
		Bus:            bus,
		Store:          &mockstore.MockStore{},
		Unmarshal:      proto.Unmarshal,
		Marshal:        proto.Marshal,
		BurialReceiver: NewBurialReceiver(), // nothing ever signals the burial
	}
	session, err := NewSession(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "giving up")
	assert.Nil(t, session)

	topic := messaging.UserConfigTopic("someuser")
	sub, ok := bus.subscribers[topic].(*userConfig)
	require.True(t, ok, "user config topic was never subscribed to")
	require.NoError(t, wizard.Publish(topic, "stale"))
	select {
	case msg := <-sub.updatesChannel:
		t.Fatalf("stale user-config subscriber still received %v after the burial timeout", msg)
	case <-time.After(200 * time.Millisecond):
	}
}

type recordingBus struct {
	*messaging.WizardBus
	subscribers map[string]messaging.Subscriber
}

func (b *recordingBus) Subscribe(topic, consumer string, s messaging.Subscriber) (messaging.Subscription, error) {
	b.subscribers[topic] = s
	return b.WizardBus.Subscribe(topic, consumer, s)
}

// opaqueCtx exposes its own done channel, so context.WithCancel cannot attach
// a child directly to the parent and has to start a goroutine that lives until
// the child context is cancelled. That makes the child's cancellation
// observable from outside NewSession.
type opaqueCtx struct {
	context.Context
	done chan struct{}
}

func newOpaqueCtx(parent context.Context) opaqueCtx {
	o := opaqueCtx{Context: parent, done: make(chan struct{})}
	return o
}

func (o opaqueCtx) Done() <-chan struct{} { return o.done }

func countGoroutinesIn(fn string) int {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), fn)
}

func TestNewSessionErrorCancelsChildContext(t *testing.T) {
	wizard, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, wizard.Start())
	defer func() { _ = wizard.Stop() }()
	bus := &failingPublishBus{WizardBus: wizard, subscribers: map[string]messaging.Subscriber{}}

	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	newCfg := func() SessionConfig {
		return SessionConfig{
			AgentName:     "testing",
			Namespace:     "acme",
			Subscriptions: []string{"testing"},
			Conn:          new(mocktransport.MockTransport),
			Bus:           bus,
			Store:         &mockstore.MockStore{},
			Unmarshal:     proto.Unmarshal,
			Marshal:       proto.Marshal,
		}
	}

	const n = 20
	before := countGoroutinesIn("context.(*cancelCtx).propagateCancel.func")
	for i := 0; i < n; i++ {
		_, err := NewSession(newOpaqueCtx(parent), newCfg())
		require.Error(t, err)
	}
	var leaked int
	for i := 0; i < 50; i++ {
		leaked = countGoroutinesIn("context.(*cancelCtx).propagateCancel.func") - before
		if leaked <= 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.LessOrEqual(t, leaked, 0, "child context goroutines are still running after %d failed NewSession calls", n)
}
