package eventd

import "github.com/sensu/sensu-go/types"

// isFlapping determines if the check is flapping, based on the TotalStateChange
// and configured thresholds
func isFlapping(event *types.Event) bool {
	if event.Check.LowFlapThreshold == 0 || event.Check.HighFlapThreshold == 0 {
		return false
	}

	// Is the check already flapping?
	if event.Check.State == types.EventFlappingState {
		return event.Check.TotalStateChange > event.Check.LowFlapThreshold
	}

	// The check was not flapping, now determine if it does now
	return event.Check.TotalStateChange >= event.Check.HighFlapThreshold
}

// state determines the check state based on whether the check is flapping and
// its status
func state(event *types.Event) {
	if flapping := isFlapping(event); flapping {
		event.Check.State = types.EventFlappingState
	} else if event.Check.Status == 0 {
		event.Check.State = types.EventPassingState
		event.Check.LastOK = event.Timestamp
	} else {
		event.Check.State = types.EventFailingState
	}
}

// totalStateChange calculates the total state change percentage for the
// history, which is later used for check state flap detection.
func totalStateChange(event *types.Event) uint32 {
	if len(event.Check.History) < 21 {
		return 0
	}

	stateChanges := 0.00
	changeWeight := 0.80
	previousStatus := event.Check.History[0].Status

	for i := 0; i < len(event.Check.History); i++ {
		if event.Check.History[i].Status != previousStatus {
			stateChanges += changeWeight
		}

		changeWeight += 0.02
		previousStatus = event.Check.History[i].Status
	}

	return uint32(float32(stateChanges) / 20 * 100)
}
