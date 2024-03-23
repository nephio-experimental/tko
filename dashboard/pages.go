package dashboard

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UpdateTextFunc func(textView *tview.TextView)
type UpdateTableFunc func(table *tview.Table)

type Details interface {
	GetTitle() string
	GetText() string
}

func (self *Application) AddTextPage(name string, title string, key rune, updateText UpdateTextFunc) {
	text := tview.NewTextView().
		SetDynamicColors(true).
		SetDoneFunc(func(key tcell.Key) {
			self.application.SetFocus(self.menu)
		})

	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(title).
		SetBlurFunc(func() {
			if self.ticker != nil {
				self.ticker.Stop()
				self.ticker = nil
			}
		})

	self.pages.AddPage(name, view, true, false)
	self.menu.AddItem(title, "", key, func() {
		self.pages.SwitchToPage(name)
		self.application.SetFocus(text)
		if updateText != nil {
			self.ticker = NewTicker(self.application, self.frequency, func() {
				updateText(text)
			})
			self.ticker.Start()
		}
	})
}

func (self *Application) AddTablePage(name string, title string, key rune, updateTable UpdateTableFunc) {
	table := tview.NewTable()

	openDetails := func(row int, column int) {
		if details, ok := table.GetCell(row, column).GetReference().(Details); ok {
			page, _ := self.pages.GetFrontPage()
			text := tview.NewTextView().
				SetDynamicColors(true).
				SetText(details.GetText()).
				SetDoneFunc(func(key tcell.Key) {
					self.pages.RemovePage("details")
					self.pages.SwitchToPage(page)
					self.application.SetFocus(table)
				})
			view := tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(text, 0, 1, true)
			view.
				SetBorder(true).
				SetTitle(details.GetTitle())
			self.pages.AddAndSwitchToPage("details", view, true)
			self.application.SetFocus(text)
		}
	}

	table.
		SetSelectable(true, true).
		Select(1, 0).
		SetDoneFunc(func(key tcell.Key) {
			self.application.SetFocus(self.menu)
		}).
		SetSelectedFunc(func(row int, column int) {
			openDetails(row, column)
		})
	table.
		SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftDoubleClick {
				row, column := table.GetOffset()
				x, y := event.Position()
				row += y - 1 // -1 to skip the header

				// Which column are we on?
				columns := table.GetColumnCount()
				for c := 0; c < columns; c++ {
					cell := table.GetCell(row, c)
					cx, _, width := cell.GetLastPosition()
					if (x >= cx) && (x <= cx+width) {
						column = c
						break
					}
				}

				openDetails(row, column)
				self.application.ForceDraw()
				return action, nil
			}
			return action, event
		})

	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(title).
		SetBlurFunc(func() {
			if self.ticker != nil {
				self.ticker.Stop()
				self.ticker = nil
			}
		})

	self.pages.AddPage(name, view, true, false)
	self.menu.AddItem(title, "", key, func() {
		self.pages.SwitchToPage(name)
		self.application.SetFocus(table)
		if updateTable != nil {
			self.ticker = NewTicker(self.application, self.frequency, func() {
				updateTable(table)
			})
			self.ticker.Start()
		}
	})
}
