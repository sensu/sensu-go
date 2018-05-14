package stager

import "context"

type Stager interface {
	// NextStageWithContext adds a new stage to the Stager.
	NextStage() Stage
	// NextStageWithContext adds a new stage to the Stager. Provided ctxParent is used as the parent context for the
	// Stage's context.
	NextStageWithContext(ctxParent context.Context) Stage
	// Shutdown iterates Stages in reverse order, cancelling their context and waiting for all started goroutines
	// to finish for each Stage.
	Shutdown()
}

func New() Stager {
	return &stager{}
}

type stager struct {
	stages []*stage
}

func (sr *stager) NextStage() Stage {
	return sr.NextStageWithContext(context.Background())
}

func (sr *stager) NextStageWithContext(ctxParent context.Context) Stage {
	ctx, cancel := context.WithCancel(ctxParent)
	st := &stage{
		ctx:    ctx,
		cancel: cancel,
	}
	sr.stages = append(sr.stages, st)
	return st
}

func (sr *stager) Shutdown() {
	for i := len(sr.stages) - 1; i >= 0; i-- {
		st := sr.stages[i]
		st.cancel()
		st.group.Wait()
	}
}
