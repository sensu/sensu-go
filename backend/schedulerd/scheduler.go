package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"strings"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

//
// tldr;
//
// - sets up up interval that runs every second
// - on interval grab all checks from state atom
// - filter checks that should be queued within the next 1000 milliseconds
// - queue filtered set of checks
//

// SchedulerInterval frequency in which we schedule checks
const SchedulerInterval = 1 * time.Second

// Scheduler ...
type Scheduler struct {
	StateManager *StateManager
	MessageBus   messaging.MessageBus

	stopping  chan struct{}
	waitGroup *sync.WaitGroup
}

// Start ...
func (schedulerPtr *Scheduler) Start() {
	schedulerPtr.stopping = make(chan struct{})

	// Wait for runloop to close and tasks to finish
	wg := &sync.WaitGroup{}
	wg.Add(1)
	schedulerPtr.waitGroup = wg

	// Start timer
	// NOTE: for consistency we use time until next second ticks past as duration
	timer := time.NewTimer(time.Duration(time.Now().UnixNano()) % SchedulerInterval)

	go func() {
		defer wg.Done()

		select {
		case <-schedulerPtr.stopping:
			return
		case <-timer.C:
			// Reset the timer to the next interval
			t := time.Now()
			timer.Reset(time.Duration(t.UnixNano()) % SchedulerInterval)

			// Grab state atom
			state := schedulerPtr.StateManager.State()
			taskCtx := &TaskContext{
				MessageBus: schedulerPtr.MessageBus,
				State:      &state,
			}

			// Iterate over every check looking for relevant
			for _, check := range state.Checks() {
				task := &Task{Check: check, Time: &t}

				// If the check should be run at this interval queue it
				if task.ShouldRun() {
					DoTaskAsync(task, taskCtx, schedulerPtr.waitGroup)
				}
			}
		}
	}()
}

// Stop ...
func (schedulerPtr *Scheduler) Stop() {
	close(schedulerPtr.stopping)
	schedulerPtr.waitGroup.Wait()
}

// Task ...
type Task struct {
	Time  *time.Time
	Check *types.CheckConfig
}

// ShouldRun determine if the task should be run at this time
func (taskPtr *Task) ShouldRun() bool {
	interval := taskPtr.Check.Interval

	// Calculate a check execution splay to ensure
	// execution is consistent between process restarts.
	sum := md5.Sum([]byte(taskPtr.Check.Name))
	splay := binary.LittleEndian.Uint64(sum[:])

	// Determine the duration of time until our the next
	// time the task needs to be run
	now := uint64(taskPtr.Time.UnixNano())
	offset := (splay - now) % uint64(interval)
	timeUntilNextRun := time.Duration(offset) / time.Nanosecond

	// If the check's interval intersects the schedulers we run the task
	return timeUntilNextRun < SchedulerInterval
}

// Run ...
func (taskPtr *Task) Run(ctx *TaskContext) {
	request := taskPtr.buildRequest(ctx)
	taskPtr.publishRequest(request, ctx)
	// TODO: What do we do w/ any errors?
}

func (taskPtr *Task) publishRequest(request *types.CheckRequest, ctx *TaskContext) error {
	var err error

	check := taskPtr.Check
	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Organization, sub)
		logger.Debugf("Sending check request for %s on topic %s", check.Name, topic)

		if pubErr := ctx.MessageBus.Publish(topic, request); err != nil {
			logger.Info("error publishing check request: ", err.Error())
			err = pubErr
		}
	}
	return err
}

// given check config fetches associated assets and builds request
func (taskPtr *Task) buildRequest(ctx *TaskContext) *types.CheckRequest {
	check := taskPtr.Check

	// ...
	request := &types.CheckRequest{}
	request.Config = check

	// Guard against iterating over assets if there are no assets associated with
	// the check in the first place.
	if len(check.RuntimeAssets) == 0 {
		return request
	}

	// Explode assets; get assets & filter out those that are irrelevant
	allAssets := ctx.State.GetAssetsInOrg(check.Organization)
	for _, asset := range allAssets {
		if assetIsRelevantToCheck(check, asset) {
			request.ExpandedAssets = append(request.ExpandedAssets, *asset)
		}
	}

	return request
}

// TaskContext ...
type TaskContext struct {
	State      *SchedulerState
	MessageBus messaging.MessageBus
}

// DoTaskAsync ...
func DoTaskAsync(task *Task, ctx *TaskContext, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		task.Run(ctx)
		wg.Done()
	}()
}

// Determine if the any of the check's runtime assets match the given assets
func assetIsRelevantToCheck(check *types.CheckConfig, asset *types.Asset) bool {
	for _, assetName := range check.RuntimeAssets {
		if strings.HasPrefix(asset.Name, assetName) {
			return true
		}
	}

	return false
}
