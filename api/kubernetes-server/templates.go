package server

import (
	contextpkg "context"
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewTemplateStore(backend backendpkg.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		Kind:        "Template",
		ListKind:    "TemplateList",
		Singular:    "template",
		Plural:      "templates",
		ObjectTyper: Scheme,

		NewResourceFunc: func() runtime.Object {
			return new(krm.Template)
		},

		NewResourceListFunc: func() runtime.Object {
			return new(krm.TemplateList)
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if krmTemplate, ok := object.(*krm.Template); ok {
				if template, err := KRMToTemplate(krmTemplate); err == nil {
					if err := store.Backend.SetTemplate(context, template); err == nil {
						return krmTemplate, nil
					} else {
						return nil, err
					}
				} else {
					return nil, backendpkg.NewBadArgumentError(err.Error())
				}
			} else {
				return nil, backendpkg.NewBadArgumentErrorf("not a Template: %T", object)
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteTemplate(context, id)
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if template, err := store.Backend.GetTemplate(context, id); err == nil {
				if krmTemplate, err := TemplateToKRM(template); err == nil {
					return &krmTemplate, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store) (runtime.Object, error) {
			var krmTemplateList krm.TemplateList
			krmTemplateList.APIVersion = APIVersion
			krmTemplateList.Kind = "TemplateList"

			if results, err := store.Backend.ListTemplates(context, backendpkg.ListTemplates{}); err == nil {
				if err := util.IterateResults(results, func(templateInfo backendpkg.TemplateInfo) error {
					if krmTemplate, err := TemplateInfoToKRM(&templateInfo); err == nil {
						krmTemplateList.Items = append(krmTemplateList.Items, krmTemplate)
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
					//{Name: "Metadata", Description: descriptions["metadata"]},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmTemplates))
			for index, krmTemplate := range krmTemplates {
				row := meta.TableRow{
					Cells: []any{krmTemplate.Name, krmTemplate.Spec.TemplateId},
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
		return nil, fmt.Errorf("unsupported type: %T", object)
	}
}

func TemplateInfoToKRM(templateInfo *backendpkg.TemplateInfo) (krm.Template, error) {
	name, err := tkoutil.ToKubernetesName(templateInfo.TemplateID)
	if err != nil {
		return krm.Template{}, err
	}

	var krmTemplate krm.Template
	krmTemplate.APIVersion = APIVersion
	krmTemplate.Kind = "Template"
	krmTemplate.Name = name
	krmTemplate.UID = types.UID("tko|template|" + templateInfo.TemplateID)
	//template.GenerateName = "tko-template-"
	//template.ResourceVersion = "123"
	//template.CreationTimestamp = meta.Now()

	templateId := templateInfo.TemplateID
	krmTemplate.Spec.TemplateId = &templateId
	krmTemplate.Spec.Metadata = templateInfo.Metadata
	krmTemplate.Status.DeploymentIds = templateInfo.DeploymentIDs

	return krmTemplate, nil
}

func TemplateToKRM(template *backendpkg.Template) (krm.Template, error) {
	if krmTemplate, err := TemplateInfoToKRM(&template.TemplateInfo); err == nil {
		krmTemplate.Spec.Package = ResourcesToKRM(template.Resources)
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
		return krm.Template{}, err
	}
}

func KRMToTemplate(krmTemplate *krm.Template) (*backendpkg.Template, error) {
	var templateId string
	var err error
	if templateId, err = tkoutil.FromKubernetesName(krmTemplate.Name); err != nil {
		return nil, err
	}

	template := backendpkg.Template{
		TemplateInfo: backendpkg.TemplateInfo{
			TemplateID: templateId,
			Metadata:   krmTemplate.Spec.Metadata,
		},
	}

	template.Resources = ResourcesFromKRM(krmTemplate.Spec.Package)

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