package types

import (
	"errors"
	"fmt"
)

// Validate returns an error if the CheckName and Subscription fields are not
// provided.
func (s *Silenced) Validate() error {
	checkNameErr := ValidateName(s.Check)
	subscriptionErr := ValidateName(s.Subscription)
	if checkNameErr != nil && subscriptionErr != nil {
		return errors.New("must provide check or subscription")
	}
	return nil
}

// FixtureSilenced returns a testing fixutre for a Silenced event struct.
func FixtureSilenced(checkName string) *Silenced {
	return &Silenced{
		Check: checkName,
	}
}

// SilencedID returns the canonical ID for a silenced entry. It returns non-nil
// error if both subscription and check are empty strings.
func SilencedID(subscription, check string) (string, error) {
	if subscription == "" && check == "" {
		return "", errors.New("no subscription or check specified")
	}
	if subscription == "" {
		subscription = "*"
	}
	if check == "" {
		check = "*"
	}
	return fmt.Sprintf("%s:%s", subscription, check), nil
}
