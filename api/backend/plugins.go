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
}

type PluginID struct {
	Type string
	util.GVK
}

func NewPluginID(type_ string, group string, version string, kind string) PluginID {
	return PluginID{
		Type: type_,
		GVK:  util.NewGVK(group, version, kind),
	}
}

func (self *Plugin) Clone() *Plugin {
	properties := make(map[string]string)
	for key, value := range self.Properties {
		properties[key] = value
	}
	return &Plugin{
		PluginID:   self.PluginID,
		Executor:   self.Executor,
		Arguments:  util.StringSetClone(self.Arguments),
		Properties: properties,
	}
}
