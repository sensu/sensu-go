package agentd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func BenchmarkSubPump(b *testing.B) {
	conn := &testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		b.Fatal(err)
	}

	st := &mockstore.MockStore{}
	st.On(
		"GetNamespace",
		mock.Anything,
		"acme",
	).Return(&types.Namespace{}, nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{"testing"},
	}
	session, err := NewSession(cfg, conn, bus, st)
	if err != nil {
		b.Fatal(err)
	}

	go func() {
		for range session.sendq {
		}
	}()

	session.wg.Add(1)
	go session.subPump()

	check := types.FixtureCheckRequest("checkity-check-check")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			session.checkChannel <- check
		}
	})
}
