package authorization

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type AuthorizationSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *AuthorizationSuite) SetupTest() {
	suite.ctx = context.TODO()
}

func (suite *AuthorizationSuite) TestHasPermission() {
	// When a rule doesn't have a permission matching the requested action
	rule := types.Rule{
		Permissions: []string{types.RulePermRead},
	}
	suite.False(hasPermission(rule, types.RulePermCreate))

	// When a rule has a permission matching the requested action
	rule = types.Rule{
		Permissions: []string{types.RulePermRead},
	}
	suite.True(hasPermission(rule, types.RulePermRead))
}

func TestRunSuites(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}
