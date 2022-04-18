package exec

import (
	"context"
	"os/exec"
	"time"
)

// TimeoutKillOnContextDone signals for a timeout on context cancellation
// and cleans up by attempting to force kill all child processes.
func TimeoutKillOnContextDone(ctx context.Context) TimeoutStrategy {
	return timeoutWithRetry{
		Timeout: ctx.Done(),
		Retries: -1,
	}
}

// TimeoutPolitelyTerminate signals for a timeout at a specified deadline.
// Cleans up by attempting to interrupt child processes a configurable number of times.
// After retrying/waiting, if the process has not exited, it will attempt to force kill the group and return.
func TimeoutPolitelyTerminate(deadline time.Time, signalRetires int, retryDelay time.Duration) TimeoutStrategy {
	timeoutCh := make(chan struct{})
	go func() {
		<-time.After(time.Until(deadline))
		close(timeoutCh)
	}()
	return timeoutWithRetry{
		Timeout: timeoutCh,
		Retries: signalRetires,
		Delay:   retryDelay,
	}
}

type timeoutWithRetry struct {
	Timeout <-chan struct{}
	Delay   time.Duration
	Retries int
}

func (d timeoutWithRetry) Signal() <-chan struct{} {
	return d.Timeout
}

func (d timeoutWithRetry) Cleanup(ctx context.Context, cmd *exec.Cmd, waitErrCh <-chan error) error {
	var retryCt int
	if d.Retries < 0 {
		select {
		case waitErr := <-waitErrCh:
			return handleWaitErr(waitErr)
		default:
			return KillProcess(cmd)
		}
	}

	if err := SignalTerminate(cmd); err != nil {
		return err
	}

	for {
		select {
		case waitErr := <-waitErrCh:
			return handleWaitErr(waitErr)
			// done
		case <-time.Tick(d.Delay):
			retryCt++
			if retryCt > d.Retries {
				return KillProcess(cmd)
			}
			if err := SignalTerminate(cmd); err != nil {
				return err
			}
		}
	}
}
