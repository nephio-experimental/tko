package server

import (
	contextpkg "context"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	backendpkg "github.com/nephio-experimental/tko/backend"
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
			var templateList krm.TemplateList
			templateList.APIVersion = APIVersion
			templateList.Kind = "TemplateList"

			if results, err := store.Backend.ListTemplates(context, backendpkg.ListTemplates{}); err == nil {
				if err := util.IterateResults(results, func(templateInfo backendpkg.TemplateInfo) error {
					if krmTemplate, err := TemplateInfoToKRM(&templateInfo); err == nil {
						templateList.Items = append(templateList.Items, krmTemplate)
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

			return &templateList, nil
		},
	}

	store.Init()
	return &store
}

func TemplateInfoToKRM(templateInfo *backendpkg.TemplateInfo) (krm.Template, error) {
	var template krm.Template

	name, err := IDToName(templateInfo.TemplateID)
	if err != nil {
		return template, err
	}

	template.APIVersion = APIVersion
	template.Kind = "Template"
	template.Name = name
	//template.GenerateName = "tko-template-"
	template.UID = types.UID("tko|template|" + templateInfo.TemplateID)
	//template.ResourceVersion = "123"
	template.CreationTimestamp = meta.Now()

	templateId := templateInfo.TemplateID
	template.Spec.TemplateId = &templateId
	template.Spec.Metadata = templateInfo.Metadata

	return template, nil
}

func TemplateToKRM(template *backendpkg.Template) (krm.Template, error) {
	return TemplateInfoToKRM(&template.TemplateInfo)
}

func KRMToTemplate(template *krm.Template) (*backendpkg.Template, error) {
	metadata := template.Spec.Metadata

	var id string
	if template.Spec.TemplateId != nil {
		id = *template.Spec.TemplateId
	}
	if id == "" {
		var err error
		if id, err = NameToID(template.Name); err != nil {
			return nil, err
		}
	}

	return &backendpkg.Template{
		TemplateInfo: backendpkg.TemplateInfo{
			TemplateID: id,
			Metadata:   metadata,
		},
	}, nil
}

/*
func nameToTemplateId(context contextpkg.Context, name string) (string, error) {
	var templateId string
	var found bool

	if results, err := self.Backend.ListTemplates(context, backendpkg.ListTemplates{
		MetadataPatterns: map[string]string{
			"kubernetes.name": name,
		},
	}); err == nil {
		if templateInfo, err := results.Next(); err == nil {
			templateId = templateInfo.TemplateID
			found = true
		} else {
			return "", errors.NewInternalError(err)
		}
		results.Release()
	} else if backendpkg.IsBadArgumentError(err) {
		return "", errors.NewBadRequest(err.Error())
	} else {
		return "", errors.NewInternalError(err)
	}

	if found {
		return templateId, nil
	} else {
		return "", errors.NewNotFound(self.groupResource, name)
	}
}
*/
