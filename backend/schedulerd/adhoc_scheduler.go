package schedulerd

import (
	"context"
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sirupsen/logrus"
)

func NewAdhocScheduler(ctx context.Context, queue queue.Client, executor *CheckExecutor) *AdhocScheduler {
	ctx, cancel := context.WithCancel(ctx)
	executor.force = true
	return &AdhocScheduler{
		queue:    queue,
		executor: executor,
		ctx:      ctx,
		cancel:   cancel,
	}
}

type AdhocScheduler struct {
	queue    queue.Client
	executor *CheckExecutor
	ctx      context.Context
	cancel   context.CancelFunc
}

func (a *AdhocScheduler) Start() {
	go a.schedule()
}

func (a *AdhocScheduler) schedule() {
	ctx := a.ctx
	defer a.cancel()
	for {
		res, err := a.queue.Reserve(ctx, adhocQueueName)
		if err != nil {
			if err == ctx.Err() {
				return
			}
			logger.WithError(err).Error("unexpected error reserving adhoc check")
			continue
		}
		item := res.Item()
		var check corev2.CheckConfig
		if err := json.Unmarshal(item.Value, &check); err != nil {
			logger.WithError(err).WithField("value", string(item.Value)).Error("error unmarshaling adhoc check")
			if ackErr := res.Ack(ctx); ackErr != nil {
				logger.WithError(ackErr).
					WithField("item_id", item.ID).
					Error("error acknowleding invalid adhoc check. potential poison record")
			}
			continue
		}
		logFields := logrus.Fields{
			"queue_item_id":   item.ID,
			"check_namespace": check.Namespace,
			"check_name":      check.Name,
		}
		logger.WithFields(logFields).Debug("attempting to schedule ad hoc check")

		if err := a.executor.processCheck(ctx, &check); err != nil {
			logger.WithError(err).WithFields(logFields).Error("error processing adhoc check request")
			if nackErr := res.Nack(ctx); nackErr != nil {
				logger.WithError(nackErr).
					WithFields(logFields).
					Error("error returning unprocessed adhoc check to queue")
			}
			continue
		}

		if err := res.Ack(ctx); err != nil {
			logger.WithError(err).
				WithFields(logFields).
				Error("error acknowledging processed adhoc check. potential double delivery")
		}
		logger.WithFields(logFields).Debug("sucesfully scheduled ad hoc check")
	}
}

func (a *AdhocScheduler) Stop() {
	a.cancel()
}
