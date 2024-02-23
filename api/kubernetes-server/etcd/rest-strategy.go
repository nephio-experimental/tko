package etcd

import (
	contextpkg "context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
)

//
// RESTStrategy
//

type RESTStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func NewRESTStrategy(objectTyper runtime.ObjectTyper) *RESTStrategy {
	return &RESTStrategy{
		ObjectTyper:   objectTyper,
		NameGenerator: names.SimpleNameGenerator,
	}
}

// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) NamespaceScoped() bool {
	return true
}

// ([rest.RESTCreateStrategy] interface)
func (self *RESTStrategy) PrepareForCreate(context contextpkg.Context, object runtime.Object) {
}

// ([rest.RESTCreateStrategy] interface)
func (self *RESTStrategy) Validate(context contextpkg.Context, object runtime.Object) field.ErrorList {
	return nil
}

// ([rest.RESTCreateStrategy] interface)
func (self *RESTStrategy) WarningsOnCreate(context contextpkg.Context, object runtime.Object) []string {
	return nil
}

// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) Canonicalize(object runtime.Object) {
}

// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) PrepareForUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) {
}

// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) ValidateUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) field.ErrorList {
	return nil
}

// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) WarningsOnUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) []string {
	return nil
}

// ([rest.RESTUpdateStrategy] interface)
func (self *RESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}
