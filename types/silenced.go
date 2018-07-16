package types

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

// Validate returns an error if the CheckName and Subscription fields are not
// provided.
func (s *Silenced) Validate() error {
	if (s.Subscription == "" && s.Check == "") || (s.Subscription == "*" && s.Check == "*") {
		return errors.New("must provide check or subscription")
	}
	if s.Subscription != "" && s.Subscription != "*" {
		if err := ValidateSubscriptionName(s.Subscription); err != nil {
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

// StartSilence returns true if the current unix timestamp is less than the begin
// timestamp.
func (s *Silenced) StartSilence(currentTime int64) bool {
	// if begin time is zero, it has not been set, so silencing can start.
	if s.Begin == 0 {
		return true
	}
	return currentTime > s.Begin
}

// FixtureSilenced returns a testing fixutre for a Silenced event struct.
func FixtureSilenced(id string) *Silenced {
	var check, subscription string

	parts := strings.Split(id, ":")

	if len(parts) == 2 {
		check = parts[1]
		subscription = parts[0]
	} else if len(parts) == 3 {
		check = parts[2]
		subscription = strings.Join(parts[0:2], ":")
	} else {
		panic("invalid silenced ID")
	}

	return &Silenced{
		ID:           id,
		Check:        check,
		Subscription: subscription,
		Organization: "default",
		Environment:  "default",
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

// URIPath returns the path component of a Silenced URI.
func (s *Silenced) URIPath() string {
	if s.ID == "" {
		s.ID, _ = SilencedID(s.Subscription, s.Check)
	}
	return fmt.Sprintf("/silenced/%s", url.PathEscape(s.ID))
}

// Get implements govaluate.Parameters
func (s *Silenced) Get(fname string) (interface{}, error) {
	strukt := reflect.Indirect(reflect.ValueOf(s))
	field := strukt.FieldByName(fname)
	if field.IsValid() {
		return reflect.Indirect(field).Interface(), nil
	}
	return nil, nil
}

// SortSilencedByPredicate can be used to sort a given collection using a given
// predicate.
func SortSilencedByPredicate(es []*Silenced, fn func(a, b *Silenced) bool) sort.Interface {
	return &silenceSorter{silences: es, byFn: fn}
}

// SortSilencedByID can be used to sort a given collection by their IDs.
func SortSilencedByID(es []*Silenced) sort.Interface {
	return SortSilencedByPredicate(es, func(a, b *Silenced) bool { return a.ID < b.ID })
}

// SortSilencedByBegin can be used to sort a given collection by their begin
// timestamp.
func SortSilencedByBegin(es []*Silenced) sort.Interface {
	return SortSilencedByPredicate(es, func(a, b *Silenced) bool { return a.Begin < b.Begin })
}

type silenceSorter struct {
	silences []*Silenced
	byFn     func(a, b *Silenced) bool
}

// Len implements sort.Interface.
func (s *silenceSorter) Len() int {
	return len(s.silences)
}

// Swap implements sort.Interface.
func (s *silenceSorter) Swap(i, j int) {
	s.silences[i], s.silences[j] = s.silences[j], s.silences[i]
}

// Less implements sort.Interface.
func (s *silenceSorter) Less(i, j int) bool {
	return s.byFn(s.silences[i], s.silences[j])
}
