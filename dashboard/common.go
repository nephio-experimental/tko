package dashboard

import (
	"github.com/gdamore/tcell/v2"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/rivo/tview"
	"github.com/tliron/go-transcribe"
)

var transcriber = transcribe.NewTranscriber().SetIndentSpaces(2)

func GetListMinSize(list *tview.List) (int, int) {
	w := 0
	l := list.GetItemCount()
	for i := 0; i < l; i++ {
		text1, text2 := list.GetItemText(i)
		w_ := len(text1)
		if w_ > w {
			w = w_
		}
		w_ = len(text2)
		if w_ > w {
			w = w_
		}
	}
	return w + 4, l
}

func SetTableHeader(table *tview.Table, headers ...string) {
	for column, header := range headers {
		table.SetCell(0, column, NewHeaderTableCell(header))
	}
}

func NewHeaderTableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true))
}

func BoolTableCell(v bool) *tview.TableCell {
	if v {
		return tview.NewTableCell("true").SetTextColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorBlack)
	} else {
		return tview.NewTableCell("false")
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
