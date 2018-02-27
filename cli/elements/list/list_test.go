package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type listDetailsRowTest struct {
	out *exWriter
	row *rowElem
}

func newListDetailsRowTest() *listDetailsRowTest {
	return &listDetailsRowTest{
		out: &exWriter{},
		row: &rowElem{
			label: "repo size",
			value: "16.2 GB",
		},
	}
}

func TestStyleLabel(t *testing.T) {
	// When a formatter is not configured default is used
	test := newListDetailsRowTest()
	test.row.label = "repo size"
	assert.Equal(t, "Repo Size: ", test.row.styleLabel())

	// When a formatter /is/ configured
	test.row.label = "repo size"
	test.row.labelStyle = func(_ string) string { return "cake" }
	assert.Equal(t, "cake", test.row.styleLabel())
}

func TestStyledLabel(t *testing.T) {
	// Returns styled label & memo val initially unset
	test := newListDetailsRowTest()
	test.row.label = "repo size"
	assert.Empty(t, test.row.labelMemo)
	assert.Equal(t, "Repo Size: ", test.row.styledLabel())

	// Result is memoized
	test.row.label = "something way diff"
	assert.NotEmpty(t, test.row.labelMemo)
	assert.Equal(t, "Repo Size: ", test.row.labelMemo)
	assert.Equal(t, "Repo Size: ", test.row.styledLabel())
}

func TestWriteListDetailsRow(t *testing.T) {
	test := newListDetailsRowTest()
	test.row.label = "repo size"
	assert.Equal(t, 11, test.row.labelLen())
}

func TestLabelLen(t *testing.T) {
	test := newListDetailsRowTest()
	test.row.label = "repo size"
	test.row.value = "12GB"

	test.row.write(test.out, 11)
	assert.Equal(t, "Repo Size: 12GB\n", test.out.result)

	test.out.Clean()
	test.row.write(test.out, 20)
	assert.Equal(t, "Repo Size:          12GB\n", test.out.result)
}

func TestFormattedValue(t *testing.T) {
	// When a formatter is not configured
	test := newListDetailsRowTest()
	test.row.value = "smrt"
	assert.Equal(t, "smrt", test.row.formattedValue())

	// When a formatter /is/ configured
	test.row.value = "smrt"
	test.row.valueFormatter = func(_ string) string { return "smart" }
	assert.Equal(t, "smart", test.row.formattedValue())
}

type listDetailsTest struct {
	out  *exWriter
	list *listElem
}

func newListDetailsTest() *listDetailsTest {
	return &listDetailsTest{
		out: &exWriter{},
		list: &listElem{
			title: "Check",
			rows: []*rowElem{
				{
					label: "repo size",
					value: "16.2 GB",
				},
			},
		},
	}
}

func TestLongestLabel(t *testing.T) {
	// When a formatter is not configured
	test := newListDetailsTest()
	test.list.rows[0].labelStyle = func(_ string) string { return "1234:" }
	assert.Equal(t, 5, test.list.longestLabel())
}

func TestWriteTitle(t *testing.T) {
	test := newListDetailsTest()
	test.list.title = "Check 'disk_full'"
	test.list.writeTitle(test.out)
	assert.NotEmpty(t, test.out.result)
	assert.Regexp(t, "^=== ", test.out.result)
	assert.Regexp(t, "disk_full", test.out.result)
}

func TestWriteRows(t *testing.T) {
	test := newListDetailsTest()
	test.list.writeRows(test.out)
	assert.NotEmpty(t, test.out.result)
}

func TestWriteListDetails(t *testing.T) {
	test := newListDetailsTest()
	test.list.write(test.out)
	assert.NotEmpty(t, test.out.result)
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
