package authorization

import (
	"context"
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
			assert.Equal(tc.Want, hasPermission(rule, tc.UserPermission))
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
			assert.Equal(tc.Want, matchesRuleType(rule, tc.Resource))
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

func TestContextCanAccessResource(t *testing.T) {
	testCases := []struct {
		TestName     string
		Resource     string
		Organization string
		Action       string
		Want         bool
	}{
		{"NoMatches", "checks", "notsensu", types.RulePermCreate, false},
		{"AllMatch", "entities", "sensu", types.RulePermRead, true},
	}
	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			assert := assert.New(t)

			rules := []types.Rule{
				{"entities", "sensu", []string{types.RulePermRead}},
			}

			roles := []types.Role{
				{"test", rules},
			}

			ctx := context.TODO()
			ctx = context.WithValue(ctx, types.OrganizationKey, tc.Organization)
			ctx = context.WithValue(ctx, ContextRoleKey, roles)
			assert.Equal(tc.Want, ContextCanAccessResource(ctx, tc.Resource, tc.Action))
		})
	}
}
