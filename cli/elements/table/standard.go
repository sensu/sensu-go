package table

import (
	"os"

	"github.com/mgutz/ansi"
	"github.com/olekukonko/tablewriter"
)

var (
	// TitleStyle can be used to format a string; suitable for titles
	TitleStyle = ansi.ColorFunc("white+bh")

	// PrimaryTextStyle can be used to format a string; suitable for emphasis
	PrimaryTextStyle = ansi.ColorFunc("blue+b")

	// CTATextStyle can be used to format a string; important text
	CTATextStyle = ansi.ColorFunc("red+b:white+h") // Call To Action
)

var stdTableWriter tablewriter.Table

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
//         return event.Entity.ID                         // Called for each row
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
func (t *Table) Render(rows []*Row) {
	// (Shallow) copy standard writer
	t.writer = &tablewriter.Table{}
	*t.writer = stdTableWriter

	t.writeColumns()
	t.writeRows(rows)
	t.writer.Render()
}

func (t *Table) writeRows(rows []*Row) {
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

func init() {
	stdTableWriter = *tablewriter.NewWriter(os.Stdout)

	// Borders are extraneous; replace with unicodes spaces.
	stdTableWriter.SetBorder(false)
	stdTableWriter.SetCenterSeparator(" ")
	stdTableWriter.SetColumnSeparator(" ")
	stdTableWriter.SetRowSeparator("─")

	// By default tablewriter wants to format headers in all caps. Rework.
	// We'll uniformly format the headers as bold and capitalized when we
	// render.
	stdTableWriter.SetAutoFormatHeaders(false)
}
