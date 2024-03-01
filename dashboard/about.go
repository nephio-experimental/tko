package dashboard

import (
	"github.com/rivo/tview"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateAbout(textView *tview.TextView) {
	if about, err := self.client.About(); err == nil {
		if s, err := transcriber.Stringify(about); err == nil {
			textView.SetText(s)
		}
	}
}
