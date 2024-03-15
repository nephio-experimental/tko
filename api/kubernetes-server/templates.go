package server

import (
	contextpkg "context"
	"fmt"
	"time"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewTemplateStore(backend backendpkg.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:          "Template",
		TypeListKind:      "TemplateList",
		TypeSingular:      "template",
		TypePlural:        "templates",
		CanCreateOnUpdate: true,
		ObjectTyper:       Scheme,

		NewObjectFunc: func() runtime.Object {
			return new(krm.Template)
		},

		NewListObjectFunc: func() runtime.Object {
			return new(krm.TemplateList)
		},

		GetFieldsFunc: func(object runtime.Object) (fields.Set, error) {
			if krmTemplate, ok := object.(*krm.Template); ok {
				fields := fields.Set{
					"metadata.name": krmTemplate.Name,
				}
				if krmTemplate.Spec.TemplateId != nil {
					fields["spec.templateId"] = *krmTemplate.Spec.TemplateId
				}
				return fields, nil
			} else {
				return nil, fmt.Errorf("not a template: %T", object)
			}
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if template, err := TemplateFromKRM(object); err == nil {
				if err := store.Backend.SetTemplate(context, template); err == nil {
					return object, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteTemplate(context, id)
		},

		PurgeFunc: func(context contextpkg.Context, store *Store) error {
			return store.Backend.PurgeTemplates(context, backendpkg.SelectTemplates{})
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if template, err := store.Backend.GetTemplate(context, id); err == nil {
				if krmTemplate, err := TemplateToKRM(template); err == nil {
					return krmTemplate, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, options *metainternalversion.ListOptions, offset uint, maxCount uint) (runtime.Object, error) {
			var krmTemplateList krm.TemplateList

			var metadataPatterns map[string]string
			var err error
			if metadataPatterns, err = ToMetadataPatterns(options); err != nil {
				return nil, err
			}
			selectionPredicate := store.NewSelectionPredicate(options, false)

			if results, err := store.Backend.ListTemplates(context, backendpkg.SelectTemplates{MetadataPatterns: metadataPatterns}, backendpkg.Window{Offset: offset, MaxCount: int(maxCount)}); err == nil {
				if err := util.IterateResults(results, func(templateInfo backendpkg.TemplateInfo) error {
					if krmTemplate, err := TemplateInfoToKRM(&templateInfo); err == nil {
						if ok, err := selectionPredicate.Matches(krmTemplate); err == nil {
							if ok {
								krmTemplateList.Items = append(krmTemplateList.Items, *krmTemplate)
							}
							return nil
						} else {
							return err
						}
					} else {
						return err
					}
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}

			krmTemplateList.APIVersion = APIVersion
			krmTemplateList.Kind = "TemplateList"
			return &krmTemplateList, nil
		},

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, withHeaders bool, withObject bool) (*meta.Table, error) {
			table := new(meta.Table)

			krmTemplates, err := ToTemplatesKRM(object)
			if err != nil {
				return nil, err
			}

			if withHeaders {
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name"},
					{Name: "TemplateID", Type: "string"},
					{Name: "Updated", Type: "string", Format: "date-time"},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmTemplates))
			for index, krmTemplate := range krmTemplates {
				var updated time.Time
				var err error
				if updated, err = FromResourceVersion(krmTemplate.ResourceVersion); err != nil {
					return nil, err
				}

				row := meta.TableRow{
					Cells: []any{
						krmTemplate.Name,
						krmTemplate.Spec.TemplateId,
						updated,
					},
				}
				if withObject {
					row.Object = runtime.RawExtension{Object: &krmTemplate}
				}
				table.Rows[index] = row
			}

			return table, nil
		},
	}

	store.Init()
	return &store
}

func ToTemplatesKRM(object runtime.Object) ([]krm.Template, error) {
	switch object_ := object.(type) {
	case *krm.TemplateList:
		return object_.Items, nil
	case *krm.Template:
		return []krm.Template{*object_}, nil
	default:
		return nil, backendpkg.NewBadArgumentErrorf("unsupported type: %T", object)
	}
}

func TemplateInfoToKRM(templateInfo *backendpkg.TemplateInfo) (*krm.Template, error) {
	name, err := tkoutil.ToKubernetesName(templateInfo.TemplateID)
	if err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var krmTemplate krm.Template
	krmTemplate.APIVersion = APIVersion
	krmTemplate.Kind = "Template"
	krmTemplate.Name = name
	krmTemplate.UID = types.UID("tko|template|" + templateInfo.TemplateID)
	krmTemplate.ResourceVersion = ToResourceVersion(templateInfo.Updated)
	krmTemplate.Labels, _ = tkoutil.ToKubernetesNames(templateInfo.Metadata)

	templateId := templateInfo.TemplateID
	krmTemplate.Spec.TemplateId = &templateId
	krmTemplate.Status.DeploymentIds = templateInfo.DeploymentIDs

	return &krmTemplate, nil
}

func TemplateToKRM(template *backendpkg.Template) (*krm.Template, error) {
	if krmTemplate, err := TemplateInfoToKRM(&template.TemplateInfo); err == nil {
		krmTemplate.Spec.Package = PackageToKRM(template.Package)
		return krmTemplate, nil
		/*
			if resourcesYaml, err := tkoutil.EncodeResources("yaml", template.Resources); err == nil {
				resourcesYaml_ := util.BytesToString(resourcesYaml)
				krmTemplate.Spec.ResourcesYaml = &resourcesYaml_
				return krmTemplate, nil
			} else {
				return krm.Template{}, err
			}
		*/
	} else {
		return nil, err
	}
}

func TemplateFromKRM(object runtime.Object) (*backendpkg.Template, error) {
	var krmTemplate *krm.Template
	var ok bool
	if krmTemplate, ok = object.(*krm.Template); !ok {
		return nil, backendpkg.NewBadArgumentErrorf("not a Template: %T", object)
	}

	var templateId string
	var err error
	if templateId, err = tkoutil.FromKubernetesName(krmTemplate.Name); err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var updated time.Time
	if updated, err = FromResourceVersion(krmTemplate.ResourceVersion); err != nil {
		return nil, err
	}

	template := backendpkg.Template{
		TemplateInfo: backendpkg.TemplateInfo{
			TemplateID: templateId,
			Updated:    updated,
		},
	}

	template.Metadata, _ = tkoutil.FromKubernetesNames(krmTemplate.Labels)
	template.Package = PackageFromKRM(krmTemplate.Spec.Package)

	/*
		if krmTemplate.Spec.ResourcesYaml != nil {
			var err error
			if template.Resources, err = tkoutil.DecodeResources("yaml", util.StringToBytes(*krmTemplate.Spec.ResourcesYaml)); err != nil {
				return nil, err
			}
		}
	*/

	return &template, nil
}
