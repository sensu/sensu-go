package schedulerd

var _ = testSubscriber{}

type testSubscriber struct {
	ch chan interface{}
}

func (ts testSubscriber) Receiver() chan<- interface{} {
	return ts.ch
}
