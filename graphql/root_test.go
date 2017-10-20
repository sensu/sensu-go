package graphqlschema

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type QueryNodeResolver struct {
	suite.Suite
	resolverSuite
}

func (t *QueryNodeResolver) TestStandard() {
	record := types.FixtureUser("bob")
	t.store().On("GetUser", record.Username).Return(record, nil).Once()

	params := t.newParams(nil)
	params.Args["id"] = fmt.Sprintf("srn:users:%s", record.Username)

	res, err := t.runResolver("Query.node", params)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *QueryNodeResolver) TestBadGlobalID() {
	params := t.newParams(nil)
	params.Args["id"] = "aws:s3:testing:lol"

	res, err := t.runResolver("Query.node", params)
	t.Empty(res)
	t.Error(err)
}

func (t *QueryNodeResolver) TestUnregisteredGlobalID() {
	params := t.newParams(nil)
	params.Args["id"] = "srn:kittens:myorg:production:rufus"

	res, err := t.runResolver("Query.node", params)
	t.Empty(res)
	t.Error(err)
}

func TestQueryType(testRunner *testing.T) {
	runSuites(testRunner, new(QueryNodeResolver))
}
