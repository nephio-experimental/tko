package backend

import (
	"github.com/nephio-experimental/tko/util"
)

//
// Plugin
//

type Plugin struct {
	PluginID
	Executor   string
	Arguments  []string
	Properties map[string]string
	Triggers   []util.GVK
}

type PluginID struct {
	Type string
	Name string
}

func NewPluginID(type_ string, name string) PluginID {
	return PluginID{
		Type: type_,
		Name: name,
	}
}

// ([fmt.Stringer] interface)
func (self *PluginID) String() string {
	return self.Type + "|" + self.Name
}

func (self *Plugin) AddTrigger(group string, version string, kind string) {
	self.Triggers = append(self.Triggers, util.NewGVK(group, version, kind))
}

func (self *Plugin) TriggersAsStrings() []string {
	strings := make([]string, len(self.Triggers))
	for index, trigger := range self.Triggers {
		strings[index] = trigger.String()
	}
	return strings
}

func (self *Plugin) Clone() *Plugin {
	return &Plugin{
		PluginID:   self.PluginID,
		Executor:   self.Executor,
		Arguments:  util.CloneStringSet(self.Arguments),
		Properties: util.CloneStringMap(self.Properties),
	}
}
