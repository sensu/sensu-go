package list

import (
	"fmt"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli/elements/globals"
	padUtf8 "github.com/willf/pad/utf8"
)

// StrTransform .. use to transform values; typically used to apply unicode colors to cells
type StrTransform func(string) string

// Config describes detail list
type Config struct {
	// Title that is displayed for the list of details
	Title string
	// TitleStyle formats the title; optional.
	TitleStyle StrTransform
	// RowFormatter transforms /all/ rows; optional.
	RowFormatter StrTransform
	// LabelStyle formats the name; optional.
	LabelStyle StrTransform
	// Rows describes all rows to display
	Rows []*Row
}

// Row of data to display
type Row struct {
	// Label of the row
	Label string
	// Value to display
	Value string
}

// Print (s) out details list given configuration data
func Print(out io.Writer, cfg *Config) error {
	listElem := newListElem(cfg)
	return listElem.write(out)
}

type listElem struct {
	title      string
	titleStyle StrTransform
	rows       []*rowElem
}

type rowElem struct {
	label          string
	labelStyle     StrTransform
	labelMemo      string
	value          string
	valueFormatter StrTransform
}

func newListElem(cfg *Config) listElem {
	writer := listElem{
		title:      cfg.Title,
		titleStyle: cfg.TitleStyle,
	}

	for _, row := range cfg.Rows {
		writer.rows = append(writer.rows, &rowElem{
			label:          row.Label,
			labelStyle:     cfg.LabelStyle,
			value:          row.Value,
			valueFormatter: cfg.RowFormatter,
		})
	}

	return writer
}

func (e *listElem) write(out io.Writer) error {
	if err := e.writeTitle(out); err != nil {
		return err
	}
	return e.writeRows(out)
}

func (e *listElem) writeTitle(out io.Writer) error {
	transformer := defaultTitleStyle
	if e.titleStyle != nil {
		transformer = e.titleStyle
	}

	_, err := fmt.Fprintln(out, transformer(e.title))
	return err
}

func (e *listElem) writeRows(out io.Writer) error {
	labelLen := e.longestLabel()

	for _, row := range e.rows {
		if err := row.write(out, labelLen); err != nil {
			return err
		}
	}
	return nil
}

func (e *listElem) longestLabel() (max int) {
	for _, row := range e.rows {
		len := row.labelLen()
		if len > max {
			max = len
		}
	}
	return
}

func (e *rowElem) write(out io.Writer, len int) error {
	_, err := fmt.Fprintf(
		out,
		"%s%s\n",
		padUtf8.Right(e.styledLabel(), len, " "),
		e.formattedValue(),
	)
	return err
}

func (e *rowElem) styleLabel() string {
	if e.labelStyle != nil {
		return e.labelStyle(e.label)
	}
	return defaultLabelStyle(e.label)
}

func (e *rowElem) styledLabel() string {
	if e.labelMemo == "" {
		e.labelMemo = e.styleLabel()
	}
	return e.labelMemo
}

func (e *rowElem) labelLen() int {
	return len(e.styledLabel())
}

func (e *rowElem) formattedValue() string {
	if e.valueFormatter != nil {
		return e.valueFormatter(e.value)
	}
	return e.value
}

// sets the title to bold and prepends '==='
func defaultTitleStyle(title string) string {
	return fmt.Sprintf("=== %s", globals.TitleStyle(title))
}

// appends a colon and a space
func defaultLabelStyle(name string) string {
	return fmt.Sprintf("%s: ", strings.Title(name))
}
