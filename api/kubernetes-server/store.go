package server

import (
	contextpkg "context"
	"io"
	"net/http"
	"net/url"

	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	backendpkg "github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	"k8s.io/apimachinery/pkg/api/errors"
	metabase "k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

//
// Store
//

const Category = "tko"

var APIVersion = krm.SchemeGroupVersion.Identifier()

type Store struct {
	Backend backendpkg.Backend
	Log     commonlog.Logger

	Kind        string
	ListKind    string
	Singular    string
	Plural      string
	Short       []string
	ObjectTyper runtime.ObjectTyper

	NewResourceFunc     func() runtime.Object
	NewResourceListFunc func() runtime.Object

	// These can return backend errors
	CreateFunc func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error)
	GetFunc    func(context contextpkg.Context, store *Store, id string) (runtime.Object, error)
	ListFunc   func(context contextpkg.Context, store *Store) (runtime.Object, error)

	groupResource  schema.GroupResource
	tableConverter rest.TableConvertor
}

func (self *Store) Init() {
	self.groupResource = krm.Resource(self.Plural)
	self.tableConverter = rest.NewDefaultTableConvertor(krm.Resource(self.Plural))
}

// Note: rest.Storage is the required interface, but there are *plenty* of additional optional ones.

// ([rest.Storage] interface)
// ([rest.Creater] interface)
// ([rest.NamedCreater] interface)
func (self *Store) New() runtime.Object {
	self.Log.Info("New")
	return self.NewResourceFunc()
}

// ([rest.Storage] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Destroy() {
	self.Log.Info("Destroy")
}

// ([rest.Scoper] interface)
// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) NamespaceScoped() bool {
	self.Log.Info("NamespaceScoped")
	return false
}

// ([rest.ShortNamesProvider] interface)
func (self *Store) ShortNames() []string {
	self.Log.Info("ShortNames")
	return self.Short
}

// ([rest.SingularNameProvider] interface)
func (self *Store) Categories() []string {
	self.Log.Info("Categories")
	return []string{Category}
}

// ([rest.CategoriesProvider] interface)
func (self *Store) GetSingularName() string {
	self.Log.Info("GetSingularName")
	return self.Singular
}

// ([rest.GroupVersionKindProvider] interface)
func (self *Store) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	self.Log.Infof("GroupVersionKind: %s", containingGV)
	return containingGV.WithKind(self.Kind)
}

// ([rest.GroupVersionAcceptor] interface)
func (self *Store) AcceptsGroupVersion(gv schema.GroupVersion) bool {
	self.Log.Infof("AcceptsGroupVersion: %s", gv)
	return (gv.Group == krmgroup.GroupName) && (gv.Version == krm.Version)
}

// ([rest.Lister] interface)
// ([rest.StandardStorage] interface)
func (self *Store) NewList() runtime.Object {
	self.Log.Info("NewList")
	return self.NewResourceListFunc()
}

// ([rest.Lister] interface)
// ([rest.StandardStorage] interface)
func (self *Store) List(context contextpkg.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	self.Log.Infof("List: %v", options)

	/*
		label := labels.Everything()
		field := fields.Everything()

		if options != nil {
			if options.LabelSelector != nil {
				label = options.LabelSelector
			}
			if options.FieldSelector != nil {
				field = options.FieldSelector
			}
		}
	*/

	//namespace, _ := request.NamespaceFrom(context)
	//self.Log.Infof("List namespace=%s", namespace)

	if list, err := self.ListFunc(context, self); err == nil {
		return list, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, errors.NewBadRequest(err.Error())
	} else {
		return nil, errors.NewInternalError(err)
	}
}

// ([rest.Lister] interface)
// ([rest.TableConvertor] interface)
// ([rest.StandardStorage] interface)
func (self *Store) ConvertToTable(context contextpkg.Context, object runtime.Object, tableOptions runtime.Object) (*meta.Table, error) {
	self.Log.Infof("ConvertToTable: %v", tableOptions)
	return self.tableConverter.ConvertToTable(context, object, tableOptions)
}

// ([rest.Getter] interface)
// ([rest.Patcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Get(context contextpkg.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	self.Log.Infof("Getter.Get: %s, %v", name, options)

	//namespace, _ := request.NamespaceFrom(context)

	id, err := NameToID(name)
	if err != nil {
		return nil, errors.NewBadRequest(err.Error())
	}

	if resource, err := self.GetFunc(context, self, id); err == nil {
		return resource, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, errors.NewBadRequest(err.Error())
	} else if backendpkg.IsNotFoundError(err) {
		return nil, errors.NewNotFound(self.groupResource, name)
	} else {
		return nil, errors.NewInternalError(err)
	}
}

// ([rest.GetterWithOptions] interface)
func (self *Store) _Get(context contextpkg.Context, name string, options runtime.Object) (runtime.Object, error) {
	self.Log.Infof("GetterWithOptions.Get: %s", name)
	return self.New(), nil
}

// ([rest.GetterWithOptions] interface)
func (self *Store) _NewGetOptions() (runtime.Object, bool, string) {
	self.Log.Info("NewGetOptions")
	return nil, false, ""
}

// ([rest.GracefulDeleter] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Delete(context contextpkg.Context, name string, deleteValidation rest.ValidateObjectFunc, options *meta.DeleteOptions) (runtime.Object, bool, error) {
	self.Log.Infof("Delete: %s, %v", name, options)
	return nil, true, nil
}

// ([rest.MayReturnFullObjectDeleter] interface)
func (self *Store) DeleteReturnsDeletedObject() bool {
	self.Log.Info("DeleteReturnsDeletedObject")
	return false
}

// ([rest.CollectionDeleter] interface)
// ([rest.StandardStorage] interface)
func (self *Store) DeleteCollection(context contextpkg.Context, deleteValidation rest.ValidateObjectFunc, options *meta.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	self.Log.Infof("DeleteCollection: %v", options)
	return nil, nil
}

// ([rest.Creater] interface)
// ([rest.CreaterUpdater] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Create(context contextpkg.Context, object runtime.Object, createValidation rest.ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error) {
	// See: https://github.com/kubernetes/apiserver/blob/bd6de43ed55ef3094738331a1264554be65c22c9/pkg/registry/generic/registry/store.go#L399

	self.Log.Infof("Creater.Create: %v", options)

	objectMeta, err := metabase.Accessor(object)
	if err != nil {
		return nil, errors.NewBadRequest(err.Error())
	}

	// Generate name if necessary
	rest.FillObjectMetaSystemFields(objectMeta)
	generateName := objectMeta.GetGenerateName()
	if (len(generateName) > 0) && (len(objectMeta.GetName()) == 0) {
		objectMeta.SetName(self.GenerateName(generateName))
	}

	// RESTCreateStrategy
	if err := rest.BeforeCreate(self, context, object); err != nil {
		return nil, err
	}

	// Validate if necessary
	if createValidation != nil {
		if err := createValidation(context, object.DeepCopyObject()); err != nil {
			return nil, err
		}
	}

	if resource, err := self.CreateFunc(context, self, object); err == nil {
		return resource, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, errors.NewBadRequest(err.Error())
	} else {
		return nil, errors.NewInternalError(err)
	}
}

// ([rest.NamedCreater] interface)
func (self *Store) _Create(context contextpkg.Context, name string, object runtime.Object, createValidation rest.ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error) {
	self.Log.Infof("NamedCreater.Create: %s, %v", name, options)
	return self.New(), nil
}

// ([rest.SubresourceObjectMetaPreserver] interface)
func (self *Store) _PreserveRequestObjectMetaSystemFieldsOnSubresourceCreate() bool {
	self.Log.Info("PreserveRequestObjectMetaSystemFieldsOnSubresourceCreate")
	return false
}

// ([rest.UpdatedObjectInfo] interface)
func (self *Store) _Preconditions() *meta.Preconditions {
	self.Log.Info("Preconditions")
	return nil
}

// ([rest.UpdatedObjectInfo] interface)
func (self *Store) UpdatedObject(context contextpkg.Context, oldObject runtime.Object) (runtime.Object, error) {
	self.Log.Info("UpdatedObject")
	return nil, nil
}

// ([rest.Updater] interface)
// ([rest.CreaterUpdater] interface)
// ([rest.Patcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Update(context contextpkg.Context, name string, objectInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	self.Log.Infof("Update: %s, %v", name, options)
	return self.New(), false, nil
}

// ([rest.Watcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) _Watch(context contextpkg.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	self.Log.Infof("Watch: %v", options)
	return nil, nil
}

// ([rest.Redirector] interface)
func (self *Store) _ResourceLocation(context contextpkg.Context, id string) (*url.URL, http.RoundTripper, error) {
	self.Log.Infof("ResourceLocation: %s", id)
	return nil, nil, nil
}

// ([rest.Responder] interface)
func (self *Store) _Object(statusCode int, obj runtime.Object) {
	self.Log.Infof("Object: %d", statusCode)
}

// ([rest.Responder] interface)
func (self *Store) _Error(err error) {
	self.Log.Info("Error")
}

// ([rest.Connecter] interface)
func (self *Store) _Connect(context contextpkg.Context, id string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	self.Log.Infof("Connect: %s", id)
	return nil, nil
}

// ([rest.Connecter] interface)
func (self *Store) _NewConnectOptions() (runtime.Object, bool, string) {
	self.Log.Info("NewConnectOptions")
	return nil, false, ""
}

// ([rest.Connecter] interface)
func (self *Store) _ConnectMethods() []string {
	self.Log.Info("ConnectMethods")
	return nil
}

// ([rest.ResourceStreamer] interface)
func (self *Store) _InputStream(context contextpkg.Context, apiVersion string, acceptHeader string) (io.ReadCloser, bool, string, error) {
	self.Log.Infof("InputStream: %s, %s", apiVersion, acceptHeader)
	return nil, false, "", nil
}

// ([rest.StorageMetadata] interface)
func (self *Store) _ProducesMIMETypes(verb string) []string {
	self.Log.Infof("ProducesMIMETypes: %s", verb)
	return nil
}

// ([rest.StorageMetadata] interface)
func (self *Store) _ProducesObject(verb string) any {
	self.Log.Infof("ProducesObject: %s", verb)
	return nil
}

// ([rest.StorageVersionProvider] interface)
func (self *Store) _StorageVersion() runtime.GroupVersioner {
	self.Log.Info("StorageVersion")
	return nil
}

// ([rest.ResetFieldsStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	self.Log.Info("GetResetFields")
	return nil
}

// ([runtime.ObjectTyper] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) ObjectKinds(object runtime.Object) ([]schema.GroupVersionKind, bool, error) {
	self.Log.Info("ObjectKinds")
	return self.ObjectTyper.ObjectKinds(object)
}

// ([runtime.ObjectTyper] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) Recognizes(gvk schema.GroupVersionKind) bool {
	self.Log.Infof("Recognizes: %s", gvk)
	return (gvk.Group == krmgroup.GroupName) && (gvk.Version == krm.Version) && (gvk.Kind == self.Kind)
}

// ([names.NameGenerator] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) GenerateName(base string) string {
	self.Log.Infof("GenerateName: %s", base)
	return names.SimpleNameGenerator.GenerateName(base)
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _AllowCreateOnUpdate() bool {
	self.Log.Info("AllowCreateOnUpdate")
	return true
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _PrepareForUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) {
	self.Log.Info("PrepareForUpdate")
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _ValidateUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) field.ErrorList {
	self.Log.Info("ValidateUpdate")
	return nil
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _WarningsOnUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) []string {
	self.Log.Info("WarningsOnUpdate")
	return nil
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) Canonicalize(object runtime.Object) {
	self.Log.Info("Canonicalize")
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) _AllowUnconditionalUpdate() bool {
	self.Log.Info("AllowUnconditionalUpdate")
	return false
}

// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
func (self *Store) PrepareForCreate(context contextpkg.Context, object runtime.Object) {
	self.Log.Info("PrepareForCreate")
}

// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
func (self *Store) Validate(context contextpkg.Context, object runtime.Object) field.ErrorList {
	self.Log.Info("Validate")
	return nil
}

// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
func (self *Store) WarningsOnCreate(context contextpkg.Context, object runtime.Object) []string {
	self.Log.Info("WarningsOnCreate")
	return nil
}
