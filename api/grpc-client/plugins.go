package client

import (
	contextpkg "context"
	"io"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/util"
)

type PluginInfo struct {
	PluginID   `json:",inline" yaml:",inline"`
	Executor   string            `json:"executor" yaml:"executor"`
	Arguments  []string          `json:"arguments" yaml:"arguments"`
	Properties map[string]string `json:"properties" yaml:"properties"`
}

type PluginID struct {
	Type     string `json:"type" yaml:"type"`
	util.GVK `json:",inline" yaml:",inline"`
}

func (self *Client) RegisterPlugin(pluginId PluginID, executor string, arguments []string, properties map[string]string) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("registerPlugin")
		if response, err := apiClient.RegisterPlugin(context, &api.Plugin{
			Type:       pluginId.Type,
			Group:      pluginId.Group,
			Version:    pluginId.Version,
			Kind:       pluginId.Kind,
			Executor:   executor,
			Arguments:  arguments,
			Properties: properties,
		}); err == nil {
			return response.Registered, response.NotRegisteredReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func NewPluginID(type_ string, gvk util.GVK) PluginID {
	return PluginID{
		Type: type_,
		GVK:  gvk,
	}
}

func (self *Client) GetPlugin(pluginId PluginID) (PluginInfo, bool, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getPlugin")
		if plugin, err := apiClient.GetPlugin(context, &api.GetPlugin{
			Type:    pluginId.Type,
			Group:   pluginId.Group,
			Version: pluginId.Version,
			Kind:    pluginId.Kind,
		}); err == nil {
			return PluginInfo{
				PluginID:   NewPluginID(plugin.Type, util.NewGVK(plugin.Group, plugin.Version, plugin.Kind)),
				Executor:   plugin.Executor,
				Arguments:  plugin.Arguments,
				Properties: plugin.Properties,
			}, true, nil
		} else if IsNotFoundError(err) {
			return PluginInfo{}, false, nil
		} else {
			return PluginInfo{}, false, err
		}
	} else {
		return PluginInfo{}, false, err
	}
}

func (self *Client) DeletePlugin(pluginId PluginID) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("deletePlugin")
		if response, err := apiClient.DeletePlugin(context, &api.DeletePlugin{
			Type:    pluginId.Type,
			Group:   pluginId.Group,
			Version: pluginId.Version,
			Kind:    pluginId.Kind,
		}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) ListPlugins() ([]PluginInfo, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("listPlugins")
		if client, err := apiClient.ListPlugins(context, new(api.ListPlugins)); err == nil {
			var plugins []PluginInfo
			for {
				if response, err := client.Recv(); err == nil {
					plugins = append(plugins, PluginInfo{
						PluginID:   NewPluginID(response.Type, util.NewGVK(response.Group, response.Version, response.Kind)),
						Executor:   response.Executor,
						Arguments:  response.Arguments,
						Properties: response.Properties,
					})
				} else if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}
			return plugins, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
