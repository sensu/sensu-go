package types

import "errors"

// Silenced returns a struct representing a silenced entry.
type Silenced struct {
	ID              string `json:"id"`
	Expire          int    `json:"expire,omitempty"`
	ExpireOnResolve bool   `json:"expire_on_resolve,omitempty"`
	Creator         string `json:"creator,omitempty"`
	CheckName       string `json:"check,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Subscription    string `json:"subscription,omitempty"`
	Organization    string `json:"organization"`
	Environment     string `json:"environment"`
}

// Validate returns an error if the CheckName and Subscription fields are not
// provided.
func (s *Silenced) Validate() error {
	checkNameErr := ValidateName(s.CheckName)
	subscriptionErr := ValidateName(s.Subscription)
	if checkNameErr != nil && subscriptionErr != nil {
		return errors.New("must provide check or subscription")
	}
	return nil
}

// FixtureSilenced returns a testing fixutre for a Silenced event struct.
func FixtureSilenced(checkName string) *Silenced {
	return &Silenced{
		CheckName: checkName,
	}
}
