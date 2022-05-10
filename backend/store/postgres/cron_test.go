package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestCronLoop(t *testing.T) {
	t.Parallel()
	ch := make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse("* * * * * *")
	if err != nil {
		t.Fatal(err)
	}
	go cronLoop(ctx, sched, ch)
	t1 := <-ch
	t2 := <-ch
	if t2.Sub(t1) < time.Millisecond || t2.Sub(t1) > 2*time.Second {
		t.Fatalf("cron loop not working: t1 %s, t2 %s", t1, t2)
	}
}
