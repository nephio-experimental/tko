package etcd

import (
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	server "github.com/nephio-experimental/tko/api/kubernetes-server"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
)

func NewTemplateStoreEtcd(restOptions generic.RESTOptionsGetter) (*registry.Store, error) {
	strategy := NewRESTStrategy(server.Scheme)

	store := registry.Store{
		NewFunc: func() runtime.Object {
			return new(krm.Template)
		},
		NewListFunc: func() runtime.Object {
			return new(krm.TemplateList)
		},
		PredicateFunc:             TemplateSelectionPredicate,
		DefaultQualifiedResource:  krm.Resource("templates"),
		SingularQualifiedResource: krm.Resource("template"),
		CreateStrategy:            strategy,
		UpdateStrategy:            strategy,
		DeleteStrategy:            strategy,
		TableConvertor:            rest.NewDefaultTableConvertor(krm.Resource("templates")),
	}

	if err := store.CompleteWithOptions(&generic.StoreOptions{
		RESTOptions: restOptions,
		AttrFunc:    GetTemplateAttrs,
	}); err == nil {
		return &store, nil
	} else {
		return nil, err
	}
}

func TemplateSelectionPredicate(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetTemplateAttrs,
	}
}

// ([storage.AttrFunc] signature)
func GetTemplateAttrs(object runtime.Object) (labels.Set, fields.Set, error) {
	if template, ok := object.(*krm.Template); ok {
		return labels.Set(template.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&template.ObjectMeta, true), nil
	} else {
		return nil, nil, fmt.Errorf("not a Template: %T", object)
	}
}
