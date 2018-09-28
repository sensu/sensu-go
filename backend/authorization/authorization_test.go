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

func TestMatchesRuleEnvironment(t *testing.T) {
	testCases := []struct {
		RuleEnvironment string
		Environment     string
		Want            bool
	}{
		{"dev", "prod", false},
		{"dev", "dev", true},
		{"*", "dev", true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("%s matches %s", tc.RuleEnvironment, tc.Environment)
		t.Run(testName, func(t *testing.T) {
			assert := assert.New(t)
			rule := types.Rule{
				Environment: tc.RuleEnvironment,
			}
			assert.Equal(tc.Want, matchesRuleEnvironment(rule, tc.Environment))
		})
	}
}

func TestMatchesRuleOrganization(t *testing.T) {
	testCases := []struct {
		RuleOrganization string
		Organization     string
		Want             bool
	}{
		{"sensu", "notsensu", false},
		{"sensu", "sensu", true},
		{"*", "sensu", true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("%s matches %s", tc.RuleOrganization, tc.Organization)
		t.Run(testName, func(t *testing.T) {
			assert := assert.New(t)
			rule := types.Rule{
				Organization: tc.RuleOrganization,
			}
			assert.Equal(tc.Want, matchesRuleOrganization(rule, tc.Organization))
		})
	}
}

func TestCanAccessResource(t *testing.T) {
	testCases := []struct {
		TestName     string
		Resource     string
		Environment  string
		Organization string
		Action       string
		Want         bool
	}{
		{"NoMatches", "checks", "prod", "notsensu", types.RulePermCreate, false},
		{"AllMatch", "entities", "dev", "sensu", types.RulePermRead, true},
		{"ReadItsEnvironment", types.RuleTypeEnvironment, "dev", "sensu", types.RulePermRead, true},
		{"ReadItsOrganization", types.RuleTypeOrganization, "", "sensu", types.RulePermRead, true},
		{"ReadAnotherEnvironment", types.RuleTypeEnvironment, "prod", "sensu", types.RulePermRead, false},
		{"ReadAnotherOrganization", types.RuleTypeOrganization, "", "acme", types.RulePermRead, false},
	}
	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			assert := assert.New(t)
			actor := Actor{
				Name: "bob",
				Rules: []types.Rule{
					{
						Type:         "entities",
						Organization: "sensu",
						Environment:  "dev",
						Permissions:  []string{types.RulePermRead},
					},
				},
			}

			assert.Equal(tc.Want, CanAccessResource(actor, tc.Organization, tc.Environment, tc.Resource, tc.Action))
		})
	}
}

func TestPrivilegeEscalation(t *testing.T) {
	assert := assert.New(t)
	actor := Actor{
		Name: "bob",
		Rules: []types.Rule{
			{
				Type:         "rules",
				Organization: "hasaccess",
				Environment:  "hasaccess",
				Permissions:  []string{types.RulePermCreate},
			},
			{
				Type:         "roles",
				Organization: "hasaccess",
				Environment:  "hasaccess",
				Permissions:  []string{types.RulePermCreate},
			},
		},
	}

	assert.Equal(
		false,
		CanAccessResource(
			actor,
			"doesnothaveaccess",
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
			"doesnothaveaccess",
			"rules",
			types.RulePermCreate,
		),
	)
}
