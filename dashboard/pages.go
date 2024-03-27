package dashboard

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UpdateTextFunc func(textView *tview.TextView) error
type UpdateTableFunc func(table *tview.Table) error

type Details interface {
	GetTitle() string
	GetText() (string, error)
}

func (self *Application) SwitchToPage(page string) {
	self.pages.SwitchToPage(page)
	if focus, ok := self.pageFocus[page]; ok {
		self.application.SetFocus(focus)
	}
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

	self.pageFocus[name] = text
}

func (self *Application) AddTablePage(name string, title string, key rune, updateTable UpdateTableFunc) {
	table := tview.NewTable()

	openDetails := func(row int, column int) bool {
		if details, ok := table.GetCell(row, column).GetReference().(Details); ok {
			if text, err := details.GetText(); err == nil {
				page, _ := self.pages.GetFrontPage()
				textView := tview.NewTextView().
					SetDynamicColors(true).
					SetText(text).
					SetDoneFunc(func(key tcell.Key) {
						self.SwitchToPage(page)
						self.pages.RemovePage("details")
					})
				self.RightClickToESC(textView.Box)
				view := tview.NewFlex().
					SetDirection(tview.FlexRow).
					AddItem(textView, 0, 1, true)
				view.
					SetBorder(true).
					SetTitle(details.GetTitle())
				self.pages.AddAndSwitchToPage("details", view, true)
				self.application.SetFocus(textView)
			} else {
				self.Error(err)
			}
			return true
		} else {
			return false
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
			switch action {
			//case tview.MouseRightDoubleClick:
			//self.Error(errors.New("Test Error"))
			//return action, nil

			case tview.MouseLeftDoubleClick:
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

				if openDetails(row, column) {
					self.application.ForceDraw()
				}
				return action, nil

			default:
				return action, event
			}
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
		self.SwitchToPage(name)
		if updateTable != nil {
			self.ticker = NewTicker(self.application, self.frequency, func() {
				updateTable(table)
			})
			self.ticker.Start()
		}
	})

	self.pageFocus[name] = table
}

func (self *Application) Error(err error) {
	page, _ := self.pages.GetFrontPage()

	modal := tview.NewModal().
		SetText(err.Error()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			self.SwitchToPage(page)
			self.pages.RemovePage("error")
		})
	self.RightClickToESC(modal.Box)

	self.pages.AddAndSwitchToPage("error", modal, true)
	self.application.SetFocus(modal)
}

func (self *Application) RightClickToESC(box *tview.Box) {
	box.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case tview.MouseRightClick:
			self.application.QueueEvent(tcell.NewEventKey(tcell.KeyESC, 0, tcell.ModNone))
			return action, nil

		default:
			return action, event
		}
	})
}
