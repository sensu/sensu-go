package eventd

import "github.com/sensu/sensu-go/types"

// totalStateChange calculates the total state change percentage for the
// history, which is later used for check state flap detection.
func totalStateChange(event *types.Event) int {
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

	return int(float64(stateChanges) / 20 * 100)
}
