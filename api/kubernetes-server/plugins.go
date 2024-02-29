package server

import (
	contextpkg "context"
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewPluginStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:     "Plugin",
		TypeListKind: "PluginList",
		TypeSingular: "plugin",
		TypePlural:   "plugins",
		ObjectTyper:  Scheme,

		NewResourceFunc: func() runtime.Object {
			return new(krm.Plugin)
		},

		NewResourceListFunc: func() runtime.Object {
			return new(krm.PluginList)
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if krmPlugin, ok := object.(*krm.Plugin); ok {
				if plugin, err := KRMToPlugin(krmPlugin); err == nil {
					if err := store.Backend.SetPlugin(context, plugin); err == nil {
						return krmPlugin, nil
					} else {
						return nil, err
					}
				} else {
					return nil, backendpkg.NewBadArgumentError(err.Error())
				}
			} else {
				return nil, backendpkg.NewBadArgumentErrorf("not a Plugin: %T", object)
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			pluginId, ok := backendpkg.ParsePluginID(id)
			if !ok {
				return fmt.Errorf("malformed plugin ID: %s", id)
			}

			return store.Backend.DeletePlugin(context, pluginId)
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			pluginId, ok := backendpkg.ParsePluginID(id)
			if !ok {
				return nil, fmt.Errorf("malformed plugin ID: %s", id)
			}

			if plugin, err := store.Backend.GetPlugin(context, pluginId); err == nil {
				if krmPlugin, err := PluginToKRM(plugin); err == nil {
					return &krmPlugin, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, offset uint, maxCount uint) (runtime.Object, error) {
			var krmPluginList krm.PluginList
			krmPluginList.APIVersion = APIVersion
			krmPluginList.Kind = "PluginList"

			if results, err := store.Backend.ListPlugins(context, backendpkg.ListPlugins{Offset: offset, MaxCount: maxCount}); err == nil {
				if err := util.IterateResults(results, func(plugin backendpkg.Plugin) error {
					if krmPlugin, err := PluginToKRM(&plugin); err == nil {
						krmPluginList.Items = append(krmPluginList.Items, krmPlugin)
						return nil
					} else {
						return err
					}
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}

			return &krmPluginList, nil
		},

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, withHeaders bool, withObject bool) (*meta.Table, error) {
			table := new(meta.Table)

			krmPlugins, err := ToPluginsKRM(object)
			if err != nil {
				return nil, err
			}

			if withHeaders {
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name"},
					{Name: "Type", Type: "string"},
					{Name: "PluginID", Type: "string"},
					{Name: "Executor", Type: "string"},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmPlugins))
			for index, krmPlugin := range krmPlugins {
				row := meta.TableRow{
					Cells: []any{
						krmPlugin.Name,
						krmPlugin.Spec.Type,
						krmPlugin.Spec.PluginID,
						krmPlugin.Spec.Executor,
					},
				}
				if withObject {
					row.Object = runtime.RawExtension{Object: &krmPlugin}
				}
				table.Rows[index] = row
			}

			return table, nil
		},
	}

	store.Init()
	return &store
}

func ToPluginsKRM(object runtime.Object) ([]krm.Plugin, error) {
	switch object_ := object.(type) {
	case *krm.PluginList:
		return object_.Items, nil
	case *krm.Plugin:
		return []krm.Plugin{*object_}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", object)
	}
}

func PluginToKRM(plugin *backendpkg.Plugin) (krm.Plugin, error) {
	pluginIdString := plugin.PluginID.String()
	name, err := tkoutil.ToKubernetesName(pluginIdString)
	if err != nil {
		return krm.Plugin{}, err
	}

	var krmPlugin krm.Plugin
	krmPlugin.APIVersion = APIVersion
	krmPlugin.Kind = "Plugin"
	krmPlugin.Name = name
	krmPlugin.UID = types.UID("tko|plugin|" + pluginIdString)

	pluginId := plugin.PluginID
	krmPlugin.Spec.Type = &pluginId.Type
	krmPlugin.Spec.PluginID = &pluginId.Name
	krmPlugin.Spec.Executor = &plugin.Executor
	krmPlugin.Spec.Arguments = plugin.Arguments
	krmPlugin.Spec.Properties = plugin.Properties
	krmPlugin.Spec.Triggers = make([]krm.Trigger, len(plugin.Triggers))
	for index, trigger := range plugin.Triggers {
		krmPlugin.Spec.Triggers[index] = krm.Trigger{
			Group:   trigger.Group,
			Version: trigger.Version,
			Kind:    trigger.Kind,
		}
	}

	return krmPlugin, nil
}

func KRMToPlugin(krmPlugin *krm.Plugin) (*backendpkg.Plugin, error) {
	var pluginId backendpkg.PluginID
	if pluginId_, err := tkoutil.FromKubernetesName(krmPlugin.Name); err == nil {
		var ok bool
		if pluginId, ok = backendpkg.ParsePluginID(pluginId_); !ok {
			return nil, fmt.Errorf("malformed plugin name: %s", pluginId_)
		}
	} else {
		return nil, err
	}

	plugin := backendpkg.Plugin{
		PluginID:   pluginId,
		Arguments:  krmPlugin.Spec.Arguments,
		Properties: krmPlugin.Spec.Properties,
	}

	if krmPlugin.Spec.Executor != nil {
		plugin.Executor = *krmPlugin.Spec.Executor
	}

	plugin.Triggers = make([]tkoutil.GVK, 0)
	for index, trigger := range krmPlugin.Spec.Triggers {
		plugin.Triggers[index] = tkoutil.GVK{
			Group:   trigger.Group,
			Version: trigger.Version,
			Kind:    trigger.Kind,
		}
	}

	return &plugin, nil
}
