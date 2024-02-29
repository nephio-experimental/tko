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
	Timezone *time.Location

	client    *clientpkg.Client
	frequency time.Duration

	application *tview.Application
	menu        *tview.List
	pages       *tview.Pages
	ticker      *Ticker
}

func NewApplication(client *clientpkg.Client, frequency time.Duration, timezone *time.Location) *Application {
	if timezone == nil {
		timezone = time.Local
	}

	self := Application{
		Timezone:    timezone,
		client:      client,
		frequency:   frequency,
		application: tview.NewApplication(),
		menu:        tview.NewList(),
		pages:       tview.NewPages(),
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

	self.AddPage("deployments", "Deployments", 'd', self.UpdateDeployments)
	self.AddPage("sites", "Sites", 's', self.UpdateSites)
	self.AddPage("templates", "Templates", 't', self.UpdateTemplates)
	self.AddPage("plugins", "Plugins", 'p', self.UpdatePlugins)
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
