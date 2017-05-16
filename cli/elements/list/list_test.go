package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ListDetailsRowSuite struct {
	suite.Suite
	out *exWriter
	row *rowElem
}

func (suite *ListDetailsRowSuite) SetupTest() {
	suite.out = &exWriter{}
	suite.row = &rowElem{
		label: "repo size",
		value: "16.2 GB",
	}
}

func (suite *ListDetailsRowSuite) TestStyleLabel() {
	// When a formatter is not configured default is used
	suite.row.label = "repo size"
	suite.Equal("Repo Size: ", suite.row.styleLabel())

	// When a formatter /is/ configured
	suite.row.label = "repo size"
	suite.row.labelStyle = func(_ string) string { return "cake" }
	suite.Equal("cake", suite.row.styleLabel())
}

func (suite *ListDetailsRowSuite) TestStyledLabel() {
	// Returns styled label & memo val initially unset
	suite.row.label = "repo size"
	suite.Empty(suite.row.labelMemo)
	suite.Equal("Repo Size: ", suite.row.styledLabel())

	// Result is memoized
	suite.row.label = "something way diff"
	suite.NotEmpty(suite.row.labelMemo)
	suite.Equal("Repo Size: ", suite.row.labelMemo)
	suite.Equal("Repo Size: ", suite.row.styledLabel())
}

func (suite *ListDetailsRowSuite) TestWrite() {
	suite.row.label = "repo size"
	suite.Equal(11, suite.row.labelLen())
}

func (suite *ListDetailsRowSuite) TestLabelLen() {
	suite.row.label = "repo size"
	suite.row.value = "12GB"

	suite.row.write(suite.out, 11)
	suite.Equal("Repo Size: 12GB\n", suite.out.result)

	suite.out.Clean()
	suite.row.write(suite.out, 20)
	suite.Equal("Repo Size:          12GB\n", suite.out.result)
}

func (suite *ListDetailsRowSuite) TestFormattedValue() {
	// When a formatter is not configured
	suite.row.value = "smrt"
	suite.Equal("smrt", suite.row.formattedValue())

	// When a formatter /is/ configured
	suite.row.value = "smrt"
	suite.row.valueFormatter = func(_ string) string { return "smart" }
	suite.Equal("smart", suite.row.formattedValue())
}

type ListDetailsSuite struct {
	suite.Suite
	out  *exWriter
	list *listElem
}

func (suite *ListDetailsSuite) SetupTest() {
	suite.out = &exWriter{}
	suite.list = &listElem{
		title: "Check",
		rows: []*rowElem{
			{
				label: "repo size",
				value: "16.2 GB",
			},
		},
	}
}

func (suite *ListDetailsSuite) TestLongestLabel() {
	// When a formatter is not configured
	suite.list.rows[0].labelStyle = func(_ string) string { return "1234:" }
	suite.Equal(5, suite.list.longestLabel())
}

func (suite *ListDetailsSuite) TestWriteTitle() {
	suite.list.title = "Check 'disk_full'"
	suite.list.writeTitle(suite.out)
	suite.NotEmpty(suite.out.result)
	suite.Regexp("^=== ", suite.out.result)
	suite.Regexp("disk_full", suite.out.result)
}

func (suite *ListDetailsSuite) TestWriteRows() {
	suite.list.writeRows(suite.out)
	suite.NotEmpty(suite.out.result)
}

func (suite *ListDetailsSuite) TestWrite() {
	suite.list.write(suite.out)
	suite.NotEmpty(suite.out.result)
}

func TestRunSuites(t *testing.T) {
	suite.Run(t, new(ListDetailsRowSuite))
	suite.Run(t, new(ListDetailsSuite))
}

func TestPrint(t *testing.T) {
	out := &exWriter{}
	config := &Config{
		Title: "Check 'disk_full'",
		Rows: []*Row{
			{
				Label: "Name",
				Value: "disk_full",
			},
			{
				Label: "Last Result",
				Value: "so full",
			},
		},
	}

	Print(out, config)
	assert.NotEmpty(t, out.result)
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return 0, nil
}
