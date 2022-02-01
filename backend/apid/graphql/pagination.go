package graphql

import (
	"math"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.OffsetPageInfoFieldResolvers = (*offsetPageInfoImpl)(nil)

type offsetContainer struct {
	Nodes    interface{}
	PageInfo offsetPageInfo
}

type offsetPageInfo struct {
	offset       int
	limit        int
	totalCount   int
	partialCount bool
}

func newOffsetContainer(offset, limit int) offsetContainer {
	container := offsetContainer{}
	container.Nodes = make([]interface{}, 0)
	container.PageInfo.offset = offset
	container.PageInfo.limit = limit
	return container
}

//
// Implement OffsetPageInfoFieldResolvers
//

type offsetPageInfoImpl struct{}

// HasNextPage implements response to request for 'hasNextPage' field.
func (*offsetPageInfoImpl) HasNextPage(p graphql.ResolveParams) (bool, error) {
	page := p.Source.(offsetPageInfo)
	return (page.offset + page.limit) < page.totalCount, nil
}

// HasPreviousPage implements response to request for 'hasPreviousPage' field.
func (*offsetPageInfoImpl) HasPreviousPage(p graphql.ResolveParams) (bool, error) {
	page := p.Source.(offsetPageInfo)
	return page.offset > 0, nil
}

// NextOffset implements response to request for 'nextOffset' field.
func (*offsetPageInfoImpl) NextOffset(p graphql.ResolveParams) (int, error) {
	page := p.Source.(offsetPageInfo)
	nextOffset := page.offset + page.limit
	if nextOffset < page.totalCount {
		return nextOffset, nil
	}
	return 0, nil
}

// PreviousOffset implements response to request for 'previousOffset' field.
func (*offsetPageInfoImpl) PreviousOffset(p graphql.ResolveParams) (int, error) {
	page := p.Source.(offsetPageInfo)
	if page.offset > 0 {
		prevOffset := page.offset - page.limit
		return clampInt(prevOffset, 0, math.MaxInt32), nil
	}
	return 0, nil
}

// TotalCount implements response to request for 'totalCount' field.
func (*offsetPageInfoImpl) TotalCount(p graphql.ResolveParams) (int, error) {
	page := p.Source.(offsetPageInfo)
	return page.totalCount, nil
}

// PartialCount implements response to request for 'totalCount' field.
func (*offsetPageInfoImpl) PartialCount(p graphql.ResolveParams) (bool, error) {
	page := p.Source.(offsetPageInfo)
	return page.partialCount, nil
}
