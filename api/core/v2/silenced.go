package v2

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// SilencedResource is the name of this resource type
	SilencedResource = "silenced"
)

// StorePrefix returns the path prefix to this resource in the store
func (s *Silenced) StorePrefix() string {
	return SilencedResource
}

// URIPath returns the path component of a silenced entry URI.
func (s *Silenced) URIPath() string {
	if s.Name == "" {
		s.Name, _ = SilencedName(s.Subscription, s.Check)
	}
	if s.Namespace == "" {
		return path.Join(URLPrefix, SilencedResource, url.PathEscape(s.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(s.Namespace), SilencedResource, url.PathEscape(s.Name))
}

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

// StartSilence returns true if the given unix timestamp is equal to or occurs
// after the Silence's start time.
//
// Deprecated: To be removed in a future release, please simply use the Begin
// field.
func (s *Silenced) StartSilence(t int64) bool {
	return t >= s.Begin
}

// Prepare prepares a silenced entry for storage
func (s *Silenced) Prepare(ctx context.Context) {
	// Populate newSilence.Name with the subscription and checkName. Substitute a
	// splat if one of the values does not exist. If both values are empty, the
	// validator will return an error when attempting to update it in the store.
	s.Name, _ = SilencedName(s.Subscription, s.Check)

	// If begin timestamp was not already provided set it to the current time.
	if s.Begin == 0 {
		s.Begin = time.Now().Unix()
	}

	// Retrieve the subject of the JWT, which represents the logged on user, in
	// order to set it as the creator of the silenced entry
	if value := ctx.Value(ClaimsKey); value != nil {
		claims, ok := value.(*Claims)
		if ok {
			s.Creator = claims.Subject
		}
	}
}

// Matches returns true if the given check name and subscription match the silence.
//
// The two properties compared, Subscription and Check, are only compared if they are
// not empty. An empty value for either of those fields is considered a wildcard,
// ie: s.Check = "foo" && s.Subscription = "" will return true for s.Matches("foo", <anything>).
func (s *Silenced) Matches(check, subscription string) bool {
	if s == nil {
		return false
	}

	if !stringsutil.InArray(s.Subscription, []string{"*", subscription}) && s.Subscription != "" {
		return false
	}

	if !stringsutil.InArray(s.Check, []string{"*", check}) && s.Check != "" {
		return false
	}

	return true
}

// NewSilenced creates a new Silenced entry.
func NewSilenced(meta ObjectMeta) *Silenced {
	return &Silenced{ObjectMeta: meta}
}

// FixtureSilenced returns a testing fixture for a Silenced event struct.
func FixtureSilenced(name string) *Silenced {
	var check, subscription string

	parts := strings.Split(name, ":")

	if len(parts) == 2 {
		check = parts[1]
		subscription = parts[0]
	} else if len(parts) == 3 {
		check = parts[2]
		subscription = strings.Join(parts[0:2], ":")
	} else {
		panic("invalid silenced name")
	}

	return &Silenced{
		Check:        check,
		Subscription: subscription,
		ObjectMeta:   NewObjectMeta(name, "default"),
	}
}

// SilencedName returns the canonical name for a silenced entry. It returns non-nil
// error if both subscription and check are empty strings.
func SilencedName(subscription, check string) (string, error) {
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

// SortSilencedByPredicate can be used to sort a given collection using a given
// predicate.
func SortSilencedByPredicate(es []*Silenced, fn func(a, b *Silenced) bool) sort.Interface {
	return &silenceSorter{silences: es, byFn: fn}
}

// SortSilencedByName can be used to sort a given collection by their names.
func SortSilencedByName(es []*Silenced) sort.Interface {
	return SortSilencedByPredicate(es, func(a, b *Silenced) bool { return a.Name < b.Name })
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

// SilencedFields returns a set of fields that represent that resource
func SilencedFields(r Resource) map[string]string {
	resource := r.(*Silenced)
	fields := map[string]string{
		"silenced.name":              resource.ObjectMeta.Name,
		"silenced.namespace":         resource.ObjectMeta.Namespace,
		"silenced.check":             resource.Check,
		"silenced.creator":           resource.Creator,
		"silenced.expire_on_resolve": strconv.FormatBool(resource.ExpireOnResolve),
		"silenced.subscription":      resource.Subscription,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "silenced.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (s *Silenced) SetNamespace(namespace string) {
	s.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (s *Silenced) SetObjectMeta(meta ObjectMeta) {
	s.ObjectMeta = meta
}

func (*Silenced) RBACName() string {
	return "silenced"
}
