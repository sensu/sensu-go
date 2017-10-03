package eventd

import "github.com/sensu/sensu-go/types"

func isFlapping(event *types.Event) bool {
	if event.Check.Config.LowFlapThreshold == 0 || event.Check.Config.HighFlapThreshold == 0 {
		return false
	}

	// Is the check already flapping?
	if event.Check.Action == types.EventFlappingAction {
		return event.Check.TotalStateChange > event.Check.Config.LowFlapThreshold
	}

	// The check was not flapping, now determine if it does now
	return event.Check.TotalStateChange >= event.Check.Config.HighFlapThreshold
}

// totalStateChange calculates the total state change percentage for the
// history, which is later used for check state flap detection.
func totalStateChange(event *types.Event) uint {
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

	return uint(float64(stateChanges) / 20 * 100)
}
