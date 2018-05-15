package slack

import (
	"fmt"
	"testing"
	"time"
)

func TestMessageTimestamp(t *testing.T) {
	ts := time.Date(2012, 11, 6, 10, 23, 42, 123456000, time.UTC)
	var msg Message
	msg.Ts = fmt.Sprintf("%d.%s", ts.Unix(), "123456")
	if msg.Timestamp().UnixNano() != ts.UnixNano() {
		t.Errorf("%v != %v", ts, msg.Timestamp())
	}
}
