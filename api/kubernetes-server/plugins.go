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
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewPluginStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:          "Plugin",
		TypeListKind:      "PluginList",
		TypeSingular:      "plugin",
		TypePlural:        "plugins",
		CanCreateOnUpdate: true,

		NewObjectFunc: func() runtime.Object {
			return new(krm.Plugin)
		},

		NewListObjectFunc: func() runtime.Object {
			return new(krm.PluginList)
		},

		GetFieldsFunc: func(object runtime.Object) (fields.Set, error) {
			if krmPlugin, ok := object.(*krm.Plugin); ok {
				fields := fields.Set{
					"metadata.name": krmPlugin.Name,
				}
				if krmPlugin.Spec.Type != nil {
					fields["spec.type"] = *krmPlugin.Spec.Type
				}
				if krmPlugin.Spec.PluginId != nil {
					fields["spec.pluginId"] = *krmPlugin.Spec.PluginId
				}
				if krmPlugin.Spec.Executor != nil {
					fields["spec.executor"] = *krmPlugin.Spec.Executor
				}
				return fields, nil
			} else {
				return nil, fmt.Errorf("not a plugin: %T", object)
			}
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if krmPlugin, ok := object.(*krm.Plugin); ok {
				if plugin, err := PluginFromKRM(krmPlugin); err == nil {
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
				return backendpkg.NewBadArgumentErrorf("malformed plugin ID: %s", id)
			}

			return store.Backend.DeletePlugin(context, pluginId)
		},

		PurgeFunc: func(context contextpkg.Context, store *Store) error {
			return store.Backend.PurgePlugins(context, backendpkg.SelectPlugins{})
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			pluginId, ok := backendpkg.ParsePluginID(id)
			if !ok {
				return nil, backendpkg.NewBadArgumentErrorf("malformed plugin ID: %s", id)
			}

			if plugin, err := store.Backend.GetPlugin(context, pluginId); err == nil {
				if krmPlugin, err := PluginToKRM(plugin); err == nil {
					return krmPlugin, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, options *internalversion.ListOptions, offset uint, maxCount uint) (runtime.Object, error) {
			var krmPluginList krm.PluginList

			id, err := IDFromListOptions(options)
			if err != nil {
				return nil, err
			}

			if id != nil {
				// Get single plugin
				pluginId, ok := backendpkg.ParsePluginID(*id)
				if !ok {
					return nil, backendpkg.NewBadArgumentErrorf("malformed plugin ID: %s", *id)
				}

				if plugin, err := store.Backend.GetPlugin(context, pluginId); err == nil {
					if krmPlugin, err := PluginToKRM(plugin); err == nil {
						krmPluginList.Items = []krm.Plugin{*krmPlugin}
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else if results, err := store.Backend.ListPlugins(context, backendpkg.SelectPlugins{}, backendpkg.Window{Offset: offset, MaxCount: int(maxCount)}); err == nil {
				if err := util.IterateResults(results, func(plugin backendpkg.Plugin) error {
					if krmPlugin, err := PluginToKRM(&plugin); err == nil {
						krmPluginList.Items = append(krmPluginList.Items, *krmPlugin)
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

			krmPluginList.APIVersion = APIVersion
			krmPluginList.Kind = "PluginList"
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
						krmPlugin.Spec.PluginId,
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
		return nil, backendpkg.NewBadArgumentErrorf("unsupported type: %T", object)
	}
}

func PluginToKRM(plugin *backendpkg.Plugin) (*krm.Plugin, error) {
	pluginIdString := plugin.PluginID.String()
	name, err := tkoutil.ToKubernetesName(pluginIdString)
	if err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var krmPlugin krm.Plugin
	krmPlugin.APIVersion = APIVersion
	krmPlugin.Kind = "Plugin"
	krmPlugin.Name = name
	krmPlugin.UID = types.UID("tko|plugin|" + pluginIdString)

	pluginId := plugin.PluginID
	krmPlugin.Spec.Type = &pluginId.Type
	krmPlugin.Spec.PluginId = &pluginId.Name
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

	return &krmPlugin, nil
}

func PluginFromKRM(object runtime.Object) (*backendpkg.Plugin, error) {
	var krmPlugin *krm.Plugin
	var ok bool
	if krmPlugin, ok = object.(*krm.Plugin); !ok {
		return nil, backendpkg.NewBadArgumentErrorf("not a Plugin: %T", object)
	}

	var pluginId backendpkg.PluginID
	if pluginId_, err := tkoutil.FromKubernetesName(krmPlugin.Name); err == nil {
		var ok bool
		if pluginId, ok = backendpkg.ParsePluginID(pluginId_); !ok {
			return nil, backendpkg.NewBadArgumentErrorf("malformed plugin name: %s", pluginId_)
		}
	} else {
		return nil, backendpkg.NewBadArgumentError(err.Error())
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
