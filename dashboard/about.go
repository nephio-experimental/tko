package dashboard

import (
	"github.com/rivo/tview"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateAbout(textView *tview.TextView) error {
	if about, err := self.client.About(); err == nil {
		if text, err := ToYAML(about); err == nil {
			textView.SetText(text)
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
