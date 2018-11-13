package table

import (
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

var (
	// TitleStyle can be used to format a string; suitable for titles
	TitleStyle = globals.TitleStyle

	// PrimaryTextStyle can be used to format a string; suitable for emphasis
	PrimaryTextStyle = globals.PrimaryTextStyle

	// CTATextStyle can be used to format a string; important text
	CTATextStyle = globals.CTATextStyle
)

// Row describes a value that will represent a row in our table
type Row struct {
	Value interface{}
}

// Column describes a standard table's column
type Column struct {
	Title           string
	CellTransformer func(data interface{}) string
	ColumnStyle     func(string) string
}

// Table describes a set of related data
type Table struct {
	Columns []*Column
	writer  *tablewriter.Table
}

// New return a new Table given columns
//
// Usage:
//
//   table := table.New([]*table.Column{
//     {
//       Title: "Source",
//       ColumnStyle: table.PrimaryTextStyle,	            // Make each row blue
//       CellTransformer: func(data interface{}) string {
//         event, _ := data.(types.Event)
//         return event.Entity.Name                         // Called for each row
//       }
//     },
//     // ...
//   })
//
//   table.Render(data)
func New(columns []*Column) *Table {
	return &Table{Columns: columns}
}

// Render renders table to STDOUT given row values
func (t *Table) Render(io io.Writer, results interface{}) {
	// (Shallow) copy standard writer
	t.writer = newWriter(io)
	t.writeColumns()
	t.writeRows(results)
	t.writer.Render()
}

func (t *Table) writeRows(results interface{}) {
	if reflect.TypeOf(results).Kind() != reflect.Slice {
		return
	}

	slice := reflect.ValueOf(results)

	rows := make([]*Row, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		rows[i] = &Row{Value: slice.Index(i).Interface()}
	}

	for _, row := range rows {
		t.writeRow(row)
	}
}

func (t *Table) writeRow(row *Row) {
	var cells []string
	for _, column := range t.Columns {
		cell := column.CellTransformer(row.Value)

		if column.ColumnStyle != nil {
			cell = column.ColumnStyle(cell)
		}

		cells = append(cells, cell)
	}

	t.writer.Append(cells)
}

func (t *Table) writeColumns() {
	var fmtTitles []string
	for _, column := range t.Columns {
		title := TitleStyle(column.Title)
		fmtTitles = append(fmtTitles, title)
	}

	t.writer.SetHeader(fmtTitles)
}

func newWriter(io io.Writer) *tablewriter.Table {
	stdTableWriter := tablewriter.NewWriter(io)

	// Borders are extraneous; replace with unicodes spaces.
	stdTableWriter.SetBorder(false)
	stdTableWriter.SetCenterSeparator(" ")
	stdTableWriter.SetColumnSeparator(" ")
	stdTableWriter.SetRowSeparator("─")
	stdTableWriter.SetAutoWrapText(false)

	// By default tablewriter wants to format headers in all caps. Rework.
	// We'll uniformly format the headers as bold and capitalized when we
	// render.
	stdTableWriter.SetAutoFormatHeaders(false)

	return stdTableWriter
}
