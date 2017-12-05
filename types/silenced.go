package types

import (
	"errors"
	"fmt"
	"strings"
)

// Validate returns an error if the CheckName and Subscription fields are not
// provided.
func (s *Silenced) Validate() error {
	if (s.Subscription == "" && s.Check == "") || (s.Subscription == "*" && s.Check == "*") {
		return errors.New("must provide check or subscription")
	}
	if s.Subscription != "" && s.Subscription != "*" {
		if err := ValidateName(s.Subscription); err != nil {
			return fmt.Errorf("Subscription %s", err)
		}
	}
	if s.Check != "" && s.Check != "*" {
		if err := ValidateName(s.Check); err != nil {
			return fmt.Errorf("Check %s", err)
		}
	}
	return nil
}

// FixtureSilenced returns a testing fixutre for a Silenced event struct.
func FixtureSilenced(id string) *Silenced {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		panic("invalid silenced ID")
	}
	return &Silenced{
		ID:           id,
		Check:        parts[1],
		Subscription: parts[0],
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
