package client

import (
	contextpkg "context"
	"fmt"

	api "github.com/nephio-experimental/tko/api/grpc"
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

func NewPluginID(type_ string, name string) PluginID {
	return PluginID{
		Type: type_,
		Name: name,
	}
}

func (self *Client) RegisterPlugin(pluginId PluginID, executor string, arguments []string, properties map[string]string, triggers []tkoutil.GVK) (bool, string, error) {
	if !tkoutil.IsValidPluginType(pluginId.Type, false) {
		return false, "", fmt.Errorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("registerPlugin")
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
	if !tkoutil.IsValidPluginType(pluginId.Type, false) {
		return Plugin{}, false, fmt.Errorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getPlugin")
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
	if !tkoutil.IsValidPluginType(pluginId.Type, false) {
		return false, "", fmt.Errorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, pluginId.Type)
	}

	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("deletePlugin")
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

type ListPlugins struct {
	Type         *string
	NamePatterns []string
	Executor     *string
	Trigger      *tkoutil.GVK
}

func (self *Client) ListPlugins(listPlugins ListPlugins) (util.Results[Plugin], error) {
	if listPlugins.Type != nil {
		if !tkoutil.IsValidPluginType(*listPlugins.Type, true) {
			return nil, fmt.Errorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, *listPlugins.Type)
		}
	}

	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Info("listPlugins")
		if client, err := apiClient.ListPlugins(context, &api.ListPlugins{
			Type:         listPlugins.Type,
			NamePatterns: listPlugins.NamePatterns,
			Executor:     listPlugins.Executor,
			Trigger:      tkoutil.TriggerToAPI(listPlugins.Trigger),
		}); err == nil {
			stream := util.NewResultsStream[Plugin](cancel)

			go func() {
				for {
					if response, err := client.Recv(); err == nil {
						stream.Send(Plugin{
							PluginID:   NewPluginID(response.Type, response.Name),
							Executor:   response.Executor,
							Arguments:  response.Arguments,
							Properties: response.Properties,
							Triggers:   tkoutil.TriggersFromAPI(response.Triggers),
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
