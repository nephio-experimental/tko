package dashboard

import (
	"time"

	"github.com/gdamore/tcell/v2"
	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
)

type UpdateTableFunc func(*tview.Table)

type GetYAMLFunc func(id any) (string, string, error)

//
// Application
//

type Application struct {
	client    *clientpkg.Client
	frequency time.Duration

	application *tview.Application
	menu        *tview.List
	pages       *tview.Pages
	views       map[string]tview.Primitive
	ticker      *Ticker
}

func NewApplication(client *clientpkg.Client, frequency time.Duration) *Application {
	self := Application{
		client:      client,
		frequency:   frequency,
		application: tview.NewApplication(),
		menu:        tview.NewList(),
		pages:       tview.NewPages(),
		views:       make(map[string]tview.Primitive),
	}

	self.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				self.application.Stop()
				return nil
			}
		}
		return event
	})

	self.menu.
		ShowSecondaryText(false).
		SetShortcutColor(tcell.ColorBlue).
		SetDoneFunc(self.application.Stop).
		SetBorder(true).
		SetTitle("TKO (PoC)")

	self.AddTableView("deployments", "Deployments", 'd', self.UpdateDeployments)
	self.AddTableView("sites", "Sites", 's', self.UpdateSites)
	self.AddTableView("templates", "Templates", 't', self.UpdateTemplates)
	self.AddTableView("plugins", "Plugins", 'p', self.UpdatePlugins)
	self.menu.AddItem("Quit", "", 'q', self.application.Stop)
	self.application.QueueEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	menuWidth, _ := GetListMinSize(self.menu)
	layout := tview.NewFlex().
		AddItem(self.menu, menuWidth+2, 0, true).
		AddItem(self.pages, 0, 1, false)

	self.application.
		SetRoot(layout, true).
		EnableMouse(true).
		SetFocus(layout)

	return &self
}

type Details interface {
	GetTitle() string
	GetText() string
}

func (self *Application) AddTableView(name string, title string, key rune, updateTable UpdateTableFunc) {
	table := tview.NewTable()

	openDetails := func(row int, column int) {
		if details, ok := table.GetCell(row, column).GetReference().(Details); ok {
			page, _ := self.pages.GetFrontPage()
			text := tview.NewTextView().
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
		Select(1, 0).
		SetDoneFunc(func(key tcell.Key) {
			table.SetSelectable(false, false)
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

	self.views[name] = view
	self.pages.AddPage(name, view, true, false)
	self.menu.AddItem(title, "", key, func() {
		self.pages.SwitchToPage(name)
		self.application.SetFocus(table)
		table.SetSelectable(true, false)
		if updateTable != nil {
			self.ticker = NewTicker(self.application, self.frequency, func() {
				updateTable(table)
			})
			self.ticker.Start()
		}
	})
}
