package server

import (
	contextpkg "context"
	"fmt"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ([api.APIServer] interface)
func (self *Server) RegisterPlugin(context contextpkg.Context, plugin *api.Plugin) (*api.RegisterResponse, error) {
	self.Log.Infof("registerPlugin: %+v", plugin)

	if !plugins.IsValidPluginType(plugin.Type, false) {
		return new(api.RegisterResponse), status.Error(codes.InvalidArgument, fmt.Sprintf("plugin type must be %s: %s", plugins.PluginTypesDescription, plugin.Type))
	}

	if err := self.Backend.SetPlugin(context, backend.NewPlugin(plugin.Type, plugin.Name, plugin.Executor, plugin.Arguments, plugin.Properties, tkoutil.TriggersFromAPI(plugin.Triggers))); err == nil {
		return &api.RegisterResponse{Registered: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.RegisterResponse{Registered: false, NotRegisteredReason: err.Error()}, nil
	} else {
		return new(api.RegisterResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeletePlugin(context contextpkg.Context, pluginId *api.PluginID) (*api.DeleteResponse, error) {
	self.Log.Infof("deletePlugin: %+v", pluginId)

	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return new(api.DeleteResponse), status.Error(codes.InvalidArgument, fmt.Sprintf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type))
	}

	if err := self.Backend.DeletePlugin(context, backend.NewPluginID(pluginId.Type, pluginId.Name)); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetPlugin(context contextpkg.Context, pluginId *api.PluginID) (*api.Plugin, error) {
	self.Log.Infof("getPlugin: %+v", pluginId)

	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return new(api.Plugin), status.Error(codes.InvalidArgument, fmt.Sprintf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type))
	}

	if plugin, err := self.Backend.GetPlugin(context, backend.NewPluginID(pluginId.Type, pluginId.Name)); err == nil {
		return &api.Plugin{
			Type:       plugin.Type,
			Name:       plugin.Name,
			Executor:   plugin.Executor,
			Arguments:  plugin.Arguments,
			Properties: plugin.Properties,
			Triggers:   tkoutil.TriggersToAPI(plugin.Triggers),
		}, nil
	} else {
		return new(api.Plugin), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) ListPlugins(listPlugins *api.ListPlugins, server api.API_ListPluginsServer) error {
	self.Log.Infof("listPlugins: %+v", listPlugins)

	if listPlugins.Type != nil {
		if !plugins.IsValidPluginType(*listPlugins.Type, true) {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("plugin type must be %s: %s", plugins.PluginTypesDescription, *listPlugins.Type))
		}
	}

	if pluginResults, err := self.Backend.ListPlugins(server.Context(), backend.ListPlugins{
		Offset:       uint(listPlugins.Offset),
		MaxCount:     uint(listPlugins.MaxCount),
		Type:         listPlugins.Type,
		NamePatterns: listPlugins.NamePatterns,
		Executor:     listPlugins.Executor,
		Trigger:      tkoutil.TriggerFromAPI(listPlugins.Trigger),
	}); err == nil {
		if err := util.IterateResults(pluginResults, func(plugin backend.Plugin) error {
			return server.Send(&api.Plugin{
				Type:       plugin.Type,
				Name:       plugin.Name,
				Executor:   plugin.Executor,
				Arguments:  plugin.Arguments,
				Properties: plugin.Properties,
				Triggers:   tkoutil.TriggersToAPI(plugin.Triggers),
			})
		}); err != nil {
			return ToGRPCError(err)
		}

		defer pluginResults.Release()
	} else {
		return ToGRPCError(err)
	}

	return nil
}
