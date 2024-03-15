package backend

import (
	"slices"
	"strings"

	"github.com/nephio-experimental/tko/util"
)

const PluginIDSeparator = "|"

//
// PluginID
//

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

func ParsePluginID(id string) (PluginID, bool) {
	if type_, name, ok := strings.Cut(id, PluginIDSeparator); ok {
		return PluginID{type_, name}, true
	} else {
		return PluginID{}, false
	}
}

// ([fmt.Stringer] interface)
func (self *PluginID) String() string {
	return self.Type + PluginIDSeparator + self.Name
}

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

func NewPlugin(type_ string, name string, executor string, arguments []string, properties map[string]string, triggers []util.GVK) *Plugin {
	if properties == nil {
		properties = make(map[string]string)
	}
	return &Plugin{
		PluginID:   NewPluginID(type_, name),
		Executor:   executor,
		Arguments:  arguments,
		Properties: properties,
		Triggers:   triggers,
	}
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
		Arguments:  util.CloneStringList(self.Arguments),
		Properties: util.CloneStringMap(self.Properties),
	}
}

func SortPlugins(plugins []Plugin) {
	slices.SortFunc(plugins, func(a Plugin, b Plugin) int {
		if type_ := strings.Compare(a.Type, b.Type); type_ == 0 {
			return strings.Compare(a.Name, b.Name)
		} else {
			return type_
		}
	})
}

//
// SelectPlugins
//

type SelectPlugins struct {
	Type         *string
	NamePatterns []string
	Executor     *string
	Trigger      *util.GVK
}
