package server

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	api "github.com/nephio-experimental/tko/grpc"
)

// api.APIServer interface
func (self *Server) RegisterPlugin(context contextpkg.Context, plugin *api.Plugin) (*api.RegisterResponse, error) {
	self.Log.Infof("registerPlugin: %s", plugin)

	if err := self.Backend.SetPlugin(&backend.Plugin{
		PluginID:   backend.NewPluginID(plugin.Type, plugin.Group, plugin.Version, plugin.Kind),
		Executor:   plugin.Executor,
		Arguments:  plugin.Arguments,
		Properties: plugin.Properties,
	}); err == nil {
		return &api.RegisterResponse{Registered: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.RegisterResponse{Registered: false, NotRegisteredReason: err.Error()}, nil
	} else {
		return new(api.RegisterResponse), ToGRPCError(err)
	}
}

// api.APIServer interface
func (self *Server) DeletePlugin(context contextpkg.Context, deletePlugin *api.DeletePlugin) (*api.DeleteResponse, error) {
	self.Log.Infof("deletePlugin: %s", deletePlugin)

	if err := self.Backend.DeletePlugin(backend.NewPluginID(deletePlugin.Type, deletePlugin.Group, deletePlugin.Version, deletePlugin.Kind)); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// api.APIServer interface
func (self *Server) GetPlugin(context contextpkg.Context, getPlugin *api.GetPlugin) (*api.Plugin, error) {
	self.Log.Infof("getPlugin: %s", getPlugin)

	if plugin, err := self.Backend.GetPlugin(backend.NewPluginID(getPlugin.Type, getPlugin.Group, getPlugin.Version, getPlugin.Kind)); err == nil {
		return &api.Plugin{
			Type:       plugin.Type,
			Group:      plugin.Group,
			Version:    plugin.Version,
			Kind:       plugin.Kind,
			Executor:   plugin.Executor,
			Arguments:  plugin.Arguments,
			Properties: plugin.Properties,
		}, nil
	} else {
		return new(api.Plugin), ToGRPCError(err)
	}
}

// api.APIServer interface
func (self *Server) ListPlugins(listPlugins *api.ListPlugins, server api.API_ListPluginsServer) error {
	self.Log.Infof("listPlugins: %s", listPlugins)

	if plugins, err := self.Backend.ListPlugins(); err == nil {
		for _, plugin := range plugins {
			if err := server.Send(&api.ListPluginsResponse{
				Type:       plugin.Type,
				Group:      plugin.Group,
				Version:    plugin.Version,
				Kind:       plugin.Kind,
				Executor:   plugin.Executor,
				Arguments:  plugin.Arguments,
				Properties: plugin.Properties,
			}); err != nil {
				return err
			}
		}
	} else {
		return ToGRPCError(err)
	}

	return nil
}
