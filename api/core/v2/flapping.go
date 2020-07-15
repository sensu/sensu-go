package v2

// isFlapping determines if the check is flapping, based on the TotalStateChange
// and configured thresholds
func isFlapping(check *Check) bool {
	if check == nil {
		return false
	}

	if check.LowFlapThreshold == 0 || check.HighFlapThreshold == 0 {
		return false
	}

	// Is the check already flapping?
	previous := check.previousOccurrence()
	if previous != nil && previous.Flapping {
		return check.TotalStateChange > check.LowFlapThreshold
	}

	// The check was not flapping, now determine if it does now
	return check.TotalStateChange >= check.HighFlapThreshold
}

// updateCheckState determines the check state based on whether the check is
// flapping, and its status
func updateCheckState(check *Check) {
	if check == nil {
		return
	}
	check.TotalStateChange = totalStateChange(check)
	if isFlapping(check) {
		check.State = EventFlappingState
	} else if check.Status == 0 {
		check.State = EventPassingState
		check.LastOK = check.Executed
	} else {
		check.State = EventFailingState
	}
}

// totalStateChange calculates the total state change percentage for the
// history, which is later used for check state flap detection.
func totalStateChange(check *Check) uint32 {
	if check == nil || len(check.History) < 21 {
		return 0
	}

	stateChanges := 0.00
	changeWeight := 0.80
	previousStatus := check.History[0].Status

	for i := 1; i <= len(check.History)-1; i++ {
		if check.History[i].Status != previousStatus {
			stateChanges += changeWeight
		}

		changeWeight += 0.02
		previousStatus = check.History[i].Status
	}

	return uint32(float32(stateChanges) / 20 * 100)
}
