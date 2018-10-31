package authorization

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	testCases := []struct {
		RulePermission string
		UserPermission string
		Want           bool
	}{
		{types.RulePermRead, types.RulePermCreate, false},
		{types.RulePermRead, types.RulePermRead, true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("%s matches %s", tc.RulePermission, tc.UserPermission)
		t.Run(testName, func(t *testing.T) {
			assert := assert.New(t)
			rule := types.Rule{
				Permissions: []string{tc.RulePermission},
			}
			assert.Equal(tc.Want, HasPermission(rule, tc.UserPermission))
		})
	}
}

func TestMatchesRuleType(t *testing.T) {
	testCases := []struct {
		RuleType string
		Resource string
		Want     bool
	}{
		{"entities", "checks", false},
		{"entities", "entities", true},
		{"*", "entities", true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("%s matches %s", tc.RuleType, tc.Resource)
		t.Run(testName, func(t *testing.T) {
			assert := assert.New(t)
			rule := types.Rule{
				Type: tc.RuleType,
			}
			assert.Equal(tc.Want, MatchesRuleType(rule, tc.Resource))
		})
	}
}

func TestMatchesRuleNamespace(t *testing.T) {
	testCases := []struct {
		RuleNamespace string
		Namespace     string
		Want          bool
	}{
		{"sensu", "notsensu", false},
		{"sensu", "sensu", true},
		{"*", "sensu", true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("%s matches %s", tc.RuleNamespace, tc.Namespace)
		t.Run(testName, func(t *testing.T) {
			assert := assert.New(t)
			rule := types.Rule{
				Namespace: tc.RuleNamespace,
			}
			assert.Equal(tc.Want, matchesRuleNamespace(rule, tc.Namespace))
		})
	}
}

func TestCanAccessResource(t *testing.T) {
	testCases := []struct {
		TestName  string
		Resource  string
		Namespace string
		Action    string
		Want      bool
	}{
		{"wrong resource", types.RuleTypeEvent, "notsensu", types.RulePermCreate, false},
		{"right resource", types.RuleTypeEntity, "sensu", types.RulePermRead, true},
		{"right namespace", types.RuleTypeNamespace, "sensu", types.RulePermRead, true},
		{"wrong namespace", types.RuleTypeNamespace, "acme", types.RulePermRead, false},
	}
	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			assert := assert.New(t)
			actor := Actor{
				Name: "bob",
				Rules: []types.Rule{
					{
						Type:        types.RuleTypeEntity,
						Namespace:   "sensu",
						Permissions: []string{types.RulePermRead},
					},
				},
			}
			assert.Equal(tc.Want, CanAccessResource(actor, tc.Namespace, tc.Resource, tc.Action))
		})
	}
}

func TestPrivilegeEscalation(t *testing.T) {
	assert := assert.New(t)
	actor := Actor{
		Name: "bob",
		Rules: []types.Rule{
			{
				Type:        "rules",
				Namespace:   "hasaccess",
				Permissions: []string{types.RulePermCreate},
			},
			{
				Type:        "roles",
				Namespace:   "hasaccess",
				Permissions: []string{types.RulePermCreate},
			},
		},
	}

	assert.Equal(
		false,
		CanAccessResource(
			actor,
			"doesnothaveaccess",
			"roles",
			types.RulePermCreate,
		),
	)

	assert.Equal(
		false,
		CanAccessResource(
			actor,
			"doesnothaveaccess",
			"rules",
			types.RulePermCreate,
		),
	)
}
