package client

import (
	contextpkg "context"
	"fmt"
	"strings"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type Plugin struct {
	PluginID   `json:",inline" yaml:",inline"`
	Executor   string            `json:"executor" yaml:"executor"`
	Arguments  []string          `json:"arguments" yaml:"arguments"`
	Properties map[string]string `json:"properties" yaml:"properties"`
	Triggers   []tkoutil.GVK     `json:"triggers" yaml:"triggers"`
}

type PluginID struct {
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

// ([fmt.Stringer] interface)
func (self PluginID) String() string {
	return "type=" + self.Type + " name=" + self.Name
}

func NewPluginID(type_ string, name string) PluginID {
	return PluginID{
		Type: type_,
		Name: name,
	}
}

func (self *Client) RegisterPlugin(pluginId PluginID, executor string, arguments []string, properties map[string]string, triggers []tkoutil.GVK) (bool, string, error) {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return false, "", fmt.Errorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("registerPlugin",
			"pluginId", pluginId,
			"executor", executor,
			"arguments", arguments,
			"properties", properties,
			"triggers", triggers)
		if response, err := apiClient.RegisterPlugin(context, &api.Plugin{
			Type:       pluginId.Type,
			Name:       pluginId.Name,
			Executor:   executor,
			Arguments:  arguments,
			Properties: properties,
			Triggers:   tkoutil.TriggersToAPI(triggers),
		}); err == nil {
			return response.Registered, response.NotRegisteredReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) GetPlugin(pluginId PluginID) (Plugin, bool, error) {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return Plugin{}, false, fmt.Errorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getPlugin",
			"pluginId", pluginId)
		if plugin, err := apiClient.GetPlugin(context, &api.PluginID{
			Type: pluginId.Type,
			Name: pluginId.Name,
		}); err == nil {
			return Plugin{
				PluginID:   NewPluginID(plugin.Type, plugin.Name),
				Executor:   plugin.Executor,
				Arguments:  plugin.Arguments,
				Properties: plugin.Properties,
				Triggers:   tkoutil.TriggersFromAPI(plugin.Triggers),
			}, true, nil
		} else if IsNotFoundError(err) {
			return Plugin{}, false, nil
		} else {
			return Plugin{}, false, err
		}
	} else {
		return Plugin{}, false, err
	}
}

func (self *Client) DeletePlugin(pluginId PluginID) (bool, string, error) {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return false, "", fmt.Errorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("deletePlugin",
			"pluginId", pluginId)
		if response, err := apiClient.DeletePlugin(context, &api.PluginID{
			Type: pluginId.Type,
			Name: pluginId.Name,
		}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type SelectPlugins struct {
	Type         *string
	NamePatterns []string
	Executor     *string
	Trigger      *tkoutil.GVK
}

// ([fmt.Stringer] interface)
func (self SelectPlugins) String() string {
	var s []string
	if self.Type != nil {
		s = append(s, "type="+*self.Type)
	}
	if len(self.NamePatterns) > 0 {
		s = append(s, "namePatterns="+stringifyStringList(self.NamePatterns))
	}
	if self.Executor != nil {
		s = append(s, "executor="+*self.Executor)
	}
	if self.Trigger != nil {
		s = append(s, "trigger="+self.Trigger.ShortString())
	}
	return strings.Join(s, " ")
}

func (self *Client) ListAllPlugins(selectPlugins SelectPlugins) util.Results[Plugin] {
	return util.CombineResults(func(offset uint) (util.Results[Plugin], error) {
		return self.ListPlugins(selectPlugins, offset, ChunkSize)
	})
}

func (self *Client) ListPlugins(selectPlugins SelectPlugins, offset uint, maxCount int) (util.Results[Plugin], error) {
	var window *api.Window
	var err error
	if window, err = newWindow(offset, maxCount); err != nil {
		return nil, err
	}

	if selectPlugins.Type != nil {
		if !plugins.IsValidPluginType(*selectPlugins.Type, true) {
			return nil, fmt.Errorf("plugin type must be %s: %s", plugins.PluginTypesDescription, *selectPlugins.Type)
		}
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Info("listPlugins",
			"selectPlugins", selectPlugins)
		if client, err := apiClient.ListPlugins(context, &api.ListPlugins{
			Window: window,
			Select: &api.SelectPlugins{
				Type:         selectPlugins.Type,
				NamePatterns: selectPlugins.NamePatterns,
				Executor:     selectPlugins.Executor,
				Trigger:      tkoutil.TriggerToAPI(selectPlugins.Trigger),
			},
		}); err == nil {
			stream := util.NewResultsStream[Plugin](cancel)

			go func() {
				for {
					if plugin, err := client.Recv(); err == nil {
						stream.Send(Plugin{
							PluginID:   NewPluginID(plugin.Type, plugin.Name),
							Executor:   plugin.Executor,
							Arguments:  plugin.Arguments,
							Properties: plugin.Properties,
							Triggers:   tkoutil.TriggersFromAPI(plugin.Triggers),
						})
					} else {
						stream.Close(err) // special handling for io.EOF
						return
					}
				}
			}()

			return stream, nil
		} else {
			cancel()
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) PurgePlugins(selectPlugins SelectPlugins) (bool, string, error) {
	if selectPlugins.Type != nil {
		if !plugins.IsValidPluginType(*selectPlugins.Type, true) {
			return false, "", fmt.Errorf("plugin type must be %s: %s", plugins.PluginTypesDescription, *selectPlugins.Type)
		}
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("purgePlugins",
			"selectPlugins", selectPlugins)
		if response, err := apiClient.PurgePlugins(context, &api.SelectPlugins{
			Type:         selectPlugins.Type,
			NamePatterns: selectPlugins.NamePatterns,
			Executor:     selectPlugins.Executor,
			Trigger:      tkoutil.TriggerToAPI(selectPlugins.Trigger),
		}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}
