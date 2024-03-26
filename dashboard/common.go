package dashboard

import (
	"github.com/gdamore/tcell/v2"
	yamllexer "github.com/goccy/go-yaml/lexer"
	yamlprinter "github.com/goccy/go-yaml/printer"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/rivo/tview"
	"github.com/tliron/go-transcribe"
)

var transcriber = transcribe.NewTranscriber().SetIndentSpaces(2)

func GetListMinSize(list *tview.List) (int, int) {
	maxWidth := 0
	count := list.GetItemCount()
	for i := 0; i < count; i++ {
		text1, text2 := list.GetItemText(i)
		width := len(text1)
		if width > maxWidth {
			maxWidth = width
		}
		width = len(text2)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth + 4, count
}

func SetTableHeader(table *tview.Table, headers ...string) {
	if table.GetRowCount() == 0 {
		for column, header := range headers {
			table.SetCell(0, column, NewHeaderTableCell(header))
		}
	}
}

func NewHeaderTableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true))
}

func NewBoolTableCell(bool_ bool) *tview.TableCell {
	if bool_ {
		return tview.NewTableCell("true").SetTextColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorBlack)
	} else {
		return tview.NewTableCell("false")
	}
}

func CleanTableRows(table *tview.Table, exists func(row int) bool) {
	row := 1
	for row < table.GetRowCount() {
		if exists(row) {
			row++
		} else {
			table.RemoveRow(row)
		}
	}
}

func PackageToYAML(package_ tkoutil.Package) (string, error) {
	return ToYAML(ToSliceAny(package_))
}

func ToYAML(content any) (string, error) {
	if s, err := transcriber.Stringify(content); err == nil {
		tokens := yamllexer.Tokenize(s)
		for _, token := range tokens {
			// TODO: this might not be good enough to escape combinations of subsequenet "[" and "]" tokens
			token.Value = tview.Escape(token.Value)
		}
		return YAMLColorPrinter.PrintTokens(tokens), nil
	} else {
		return "", err
	}
}

// To force the transcriber to transcribe package as multiple YAML documents
func ToSliceAny(package_ tkoutil.Package) []any {
	slice := make([]any, len(package_))
	for index, resource := range package_ {
		slice[index] = resource
	}
	return slice
}

var YAMLColorPrinter = yamlprinter.Printer{
	String: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[blue]",
			Suffix: "[white]",
		}
	},
	Number: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[fuchsia]",
			Suffix: "[white]",
		}
	},
	Bool: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[teal]",
			Suffix: "[white]",
		}
	},

	MapKey: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[green]",
			Suffix: "[white]",
		}
	},

	Anchor: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[red]",
			Suffix: "[white]",
		}
	},
	Alias: func() *yamlprinter.Property {
		return &yamlprinter.Property{
			Prefix: "[yellow]",
			Suffix: "[white]",
		}
	},
}
