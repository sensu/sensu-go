package relay

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Connection describes edges and information about connection
type Connection struct {
	Edges      []*Edge
	PageInfo   PageInfo
	TotalCount int
}

// An Edge in a connection
type Edge struct {
	Node   interface{}
	Cursor Cursor
}

// PageInfo describes information about pagination in a connection.
type PageInfo struct {
	StartCursor Cursor
	EndCursor   Cursor

	HasNextPage     bool
	HasPreviousPage bool
}

// A Cursor for use in pagination
type Cursor fmt.Stringer

// Re-derives the offset from the cursor string.
func cursorToOffset(cursor string) (int, error) {
	str := ""
	b, err := base64.StdEncoding.DecodeString(string(cursor))
	if err == nil {
		str = string(b)
	}
	str = strings.Replace(str, arrayCursorPrefix, "", -1)
	offset, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.New("Invalid cursor")
	}
	return offset, nil
}

func getOffsetWithDefault(cursor string, defaultOffset int) int {
	if cursor == "" {
		return defaultOffset
	}
	offset, err := cursorToOffset(cursor)
	if err != nil {
		return defaultOffset
	}
	return offset
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func ternaryMax(a, b, c int) int {
	return max(max(a, b), c)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func ternaryMin(a, b, c int) int {
	return min(min(a, b), c)
}
