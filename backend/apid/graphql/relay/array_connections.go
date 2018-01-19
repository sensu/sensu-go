package relay

import (
	"encoding/base64"
	"fmt"
)

const arrayCursorPrefix = "--"

// NewArrayConnectionEdge instantiates new edge for use with array connections.
func NewArrayConnectionEdge(node interface{}, i int) *Edge {
	return &Edge{Node: node, Cursor: arrCursor(i)}
}

type arrCursor int

func (c arrCursor) String() string {
	str := fmt.Sprintf("%s%d", arrayCursorPrefix, c)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// ArrayConnection describes edges of connection
type ArrayConnection struct {
	ArrayConnectionInfo
	Edges []*Edge
}

// ArrayConnectionInfo holds details of current page.
type ArrayConnectionInfo struct {
	PageInfo   PageInfo
	TotalCount int

	Begin int
	End   int
}

// NewArrayConnectionInfo ... Given a slice (subset) of an array, returns a
// connection object for use in GraphQL.
//
// This function is similar to `ConnectionFromArray`, but is intended for use
// cases where you know the cardinality of the connection, consider it too large
// to materialize the entire array, and instead wish pass in a slice of the
// total result large enough to cover the range specified in `args`.
//
// Adapted from:
// https://github.com/graphql-go/relay/blob/master/array_connection.go#L44-L117
func NewArrayConnectionInfo(
	sliceStart,
	totalLen,
	first, last int,
	before, after string,
) ArrayConnectionInfo {
	sliceEnd := sliceStart + totalLen
	beforeOffset := getOffsetWithDefault(before, totalLen)
	afterOffset := getOffsetWithDefault(after, -1)

	startOffset := ternaryMax(sliceStart-1, afterOffset, -1) + 1
	endOffset := ternaryMin(sliceEnd, beforeOffset, totalLen)

	if first > 0 {
		endOffset = min(endOffset, startOffset+first)
	} else if last > 0 {
		startOffset = max(startOffset, endOffset-last)
	}

	begin := max(startOffset-sliceStart, 0)
	end := totalLen - (sliceEnd - endOffset)
	if begin > end {
		begin, end = 0, 0
	}

	var firstEdgeCursor, lastEdgeCursor Cursor
	if end > begin {
		firstEdgeCursor = arrCursor(sliceStart + begin)
		lastEdgeCursor = arrCursor(sliceStart + end)
	}

	lowerBound := 0
	if len(after) > 0 {
		lowerBound = afterOffset + 1
	}

	upperBound := totalLen
	if len(before) > 0 {
		upperBound = beforeOffset
	}

	hasPreviousPage := false
	hasNextPage := false
	if last > 0 {
		hasPreviousPage = startOffset > lowerBound
	} else if first > 0 {
		hasNextPage = endOffset < upperBound
	}

	conn := ArrayConnectionInfo{Begin: begin, End: end, TotalCount: totalLen}
	conn.PageInfo = PageInfo{
		StartCursor:     firstEdgeCursor,
		EndCursor:       lastEdgeCursor,
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     hasNextPage,
	}

	return conn
}
