package server

import (
	contextpkg "context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metabase "k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

//
// Store
//

const (
	Category           = "tko"
	ParallelBufferSize = 1000
	ParallelWorkers    = 10
)

var APIVersion = krm.SchemeGroupVersion.Identifier()

type Store struct {
	Backend backendpkg.Backend
	Log     commonlog.Logger

	TypeKind          string
	TypeListKind      string
	TypeSingular      string
	TypePlural        string
	TypeShortNames    []string
	CanCreateOnUpdate bool
	ObjectTyper       runtime.ObjectTyper

	NewObjectFunc     func() runtime.Object
	NewListObjectFunc func() runtime.Object

	// These can return backend errors
	CreateFunc func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error)
	UpdateFunc func(context contextpkg.Context, store *Store, updatedObject runtime.Object) (runtime.Object, error) // optional
	DeleteFunc func(context contextpkg.Context, store *Store, id string) error
	GetFunc    func(context contextpkg.Context, store *Store, id string) (runtime.Object, error)
	ListFunc   func(context contextpkg.Context, store *Store, offset uint, maxCount uint) (runtime.Object, error)
	TableFunc  func(context contextpkg.Context, store *Store, object runtime.Object, withHeaders bool, withObject bool) (*meta.Table, error)

	groupResource schema.GroupResource
}

func (self *Store) Init() {
	self.groupResource = krm.Resource(self.TypePlural)
}

// Note: rest.Storage is the required interface, but there are *plenty* of additional optional ones.
// We've tried to specify *all* possible functions here from all optional interfaces. Those currently
// unused are "disabled" by an underscore prefix.
//
// Also note that all functions must returns errors from "k8s.io/apimachinery/pkg/api/errors".
//
// For an example implementation on top of etcd, see:
//   https://github.com/kubernetes/apiserver/blob/v0.29.2/pkg/registry/generic/registry/store.go

var (
	validStore                               = new(Store)
	_          rest.Storage                  = validStore
	_          rest.Scoper                   = validStore
	_          rest.KindProvider             = validStore
	_          rest.ShortNamesProvider       = validStore
	_          rest.CategoriesProvider       = validStore
	_          rest.SingularNameProvider     = validStore
	_          rest.GroupVersionKindProvider = validStore
	_          rest.GroupVersionAcceptor     = validStore
	_          rest.Lister                   = validStore
	_          rest.Getter                   = validStore
	// _ rest.GetterWithOptions = validStore
	_ rest.TableConvertor             = validStore
	_ rest.GracefulDeleter            = validStore
	_ rest.MayReturnFullObjectDeleter = validStore
	_ rest.CollectionDeleter          = validStore
	_ rest.Creater                    = validStore
	// _ rest.NamedCreater = validStore
	_ rest.SubresourceObjectMetaPreserver = validStore
	_ rest.Updater                        = validStore
	_ rest.CreaterUpdater                 = validStore
	_ rest.Patcher                        = validStore
	// _ rest.Watcher = validStore
	// _ rest.StandardStorage = validStore
	// _ rest.Redirector = validStore
	// _ rest.Responder = validStore
	// _ rest.Connecter = validStore
	// _ rest.ResourceStreamer = validStore
	// _ rest.StorageMetadata = validStore
	// _ rest.StorageVersionProvider = validStore
	_ rest.ResetFieldsStrategy             = validStore
	_ rest.CreateUpdateResetFieldsStrategy = validStore
	_ rest.UpdateResetFieldsStrategy       = validStore
)

// ([rest.Storage] interface)
// ([rest.Creater] interface)
// ([rest.NamedCreater] interface)
func (self *Store) New() runtime.Object {
	self.Log.Info("New")
	return self.NewObjectFunc()
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

// ([rest.KindProvider] interface)
func (self *Store) Kind() string {
	self.Log.Info("Kind")
	return self.TypeKind
}

// ([rest.ShortNamesProvider] interface)
func (self *Store) ShortNames() []string {
	self.Log.Info("ShortNames")
	return self.TypeShortNames
}

// ([rest.SingularNameProvider] interface)
func (self *Store) Categories() []string {
	self.Log.Info("Categories")
	return []string{Category}
}

// ([rest.CategoriesProvider] interface)
func (self *Store) GetSingularName() string {
	self.Log.Info("GetSingularName")
	return self.TypeSingular
}

// ([rest.GroupVersionKindProvider] interface)
func (self *Store) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	self.Log.Infof("GroupVersionKind: containingGV=%s", containingGV)
	return containingGV.WithKind(self.TypeKind)
}

// ([rest.GroupVersionAcceptor] interface)
func (self *Store) AcceptsGroupVersion(gv schema.GroupVersion) bool {
	self.Log.Infof("AcceptsGroupVersion: gv=%s", gv)
	return (gv.Group == krmgroup.GroupName) && (gv.Version == krm.Version)
}

// ([rest.Lister] interface)
// ([rest.StandardStorage] interface)
func (self *Store) NewList() runtime.Object {
	self.Log.Info("NewList")
	return self.NewListObjectFunc()
}

// ([rest.Lister] interface)
// ([rest.StandardStorage] interface)
func (self *Store) List(context contextpkg.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	if options == nil {
		self.Log.Info("List")
	} else {
		self.Log.Infof("List: options=%+v", *options)
	}

	var offset uint
	var maxCount uint

	if options != nil {
		if options.Watch {
			return nil, apierrors.NewBadRequest("\"watch\" is not supported")
		}

		if options.Continue != "" {
			// Note: Every call results in a new query to the backend (we are not setting
			// ResourceVersion), so there is no guarantee that we are indeed continuing the
			// same result set. Thus it may be possible that certain names may repeat more
			// than once for a client that is concatenating result chunks. Clients concerned
			// about duplicates should thus do their own de-duping.
			if offset_, err := strconv.ParseUint(options.Continue, 10, 64); err == nil {
				offset = uint(offset_)
			} else {
				return nil, apierrors.NewBadRequest(fmt.Sprintf("malformed \"continue\": %s", err.Error()))
			}
		}

		maxCount = uint(options.Limit)

		if options.TimeoutSeconds != nil {
			var cancel contextpkg.CancelFunc
			context, cancel = contextpkg.WithTimeout(context, time.Duration(*options.TimeoutSeconds)*time.Second)
			defer cancel()
		}
	}

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

	if list, err := self.ListFunc(context, self, offset, maxCount); err == nil {
		// Check if there are potentially more results
		if objects, err := metabase.ExtractList(list); err == nil {
			count := uint(len(objects))

			// This should not happen, but just in case
			if count > maxCount {
				self.Log.Warningf("List: fetched too many objects: %d > %d", count, maxCount)
				metabase.SetList(list, objects[:maxCount])
				count = maxCount
			}

			if count == maxCount {
				// Note: If all the results fit exactly in maxCount then we don't actually have more results
				// and there's no need to continue, but optimizing for that case here would be challenging.
				// Ssome backends, such as SQL, might not report additional results within the query window.
				if list_, err := metabase.ListAccessor(list); err == nil {
					continue_ := strconv.FormatUint(uint64(offset+count), 10)
					self.Log.Infof("List: setting continue: %s", continue_)
					list_.SetContinue(continue_)
				} else {
					return nil, apierrors.NewInternalError(err)
				}
			}
		} else {
			return nil, apierrors.NewInternalError(err)
		}

		return list, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, apierrors.NewBadRequest(err.Error())
	} else {
		return nil, apierrors.NewInternalError(err)
	}
}

// ([rest.TableConvertor] interface)
// ([rest.Lister] interface)
// ([rest.StandardStorage] interface)
func (self *Store) ConvertToTable(context contextpkg.Context, object runtime.Object, options runtime.Object) (*meta.Table, error) {
	if options == nil {
		self.Log.Info("ConvertToTable")
	} else {
		self.Log.Infof("ConvertToTable: options=%+v", options)
	}

	tableOptions, _ := options.(*meta.TableOptions)
	withHeaders := (options == nil) || !tableOptions.NoHeaders
	withObject := (options == nil) || (tableOptions.IncludeObject != meta.IncludeNone)

	if table, err := self.TableFunc(context, self, object, withHeaders, withObject); err == nil {
		if list_, err := metabase.ListAccessor(object); err == nil {
			// Copy properties from list to table
			// ("kubectl get" expects this, but "kubectl get -o yaml" doesn't... welp!)
			table.ResourceVersion = list_.GetResourceVersion()
			table.RemainingItemCount = list_.GetRemainingItemCount()
			table.Continue = list_.GetContinue()
		}
		return table, nil
	} else {
		return nil, apierrors.NewInternalError(err)
	}
}

// ([rest.Getter] interface)
// ([rest.Patcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Get(context contextpkg.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	if options == nil {
		self.Log.Infof("Getter.Get: name=%s", name)
	} else {
		self.Log.Infof("Getter.Get: name=%s options=%+v", name, *options)
	}

	id, err := tkoutil.FromKubernetesName(name)
	if err != nil {
		return nil, apierrors.NewBadRequest(err.Error())
	}

	if object, err := self.GetFunc(context, self, id); err == nil {
		return object, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, apierrors.NewBadRequest(err.Error())
	} else if backendpkg.IsNotFoundError(err) {
		return nil, apierrors.NewNotFound(self.groupResource, name)
	} else {
		return nil, apierrors.NewInternalError(err)
	}
}

// ([rest.GetterWithOptions] interface)
func (self *Store) _Get(context contextpkg.Context, name string, options runtime.Object) (runtime.Object, error) {
	self.Log.Infof("GetterWithOptions.Get: name=%s", name)
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
	if options == nil {
		self.Log.Infof("Delete: name=%s", name)
	} else {
		self.Log.Infof("Delete: name=%s options=%+v", name, *options)
	}

	id, err := tkoutil.FromKubernetesName(name)
	if err != nil {
		return nil, false, apierrors.NewBadRequest(err.Error())
	}

	// Older clients use nil to mean no grace period
	// (for completion; we always delete immediately)
	if options == nil {
		options = meta.NewDeleteOptions(0)
	}

	/*
		// Validate if necessary
		if deleteValidation != nil {
			if err := deleteValidation(context, object.DeepCopyObject()); err != nil {
				return nil, false, err
			}
		}
	*/

	if (options != nil) && (len(options.DryRun) > 0) {
		return nil, false, nil
	}

	if err := self.DeleteFunc(context, self, id); err == nil {
		return nil, true, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, false, apierrors.NewBadRequest(err.Error())
	} else if backendpkg.IsNotFoundError(err) {
		return nil, false, apierrors.NewNotFound(self.groupResource, name)
	} else {
		return nil, false, apierrors.NewInternalError(err)
	}
}

// ([rest.MayReturnFullObjectDeleter] interface)
func (self *Store) DeleteReturnsDeletedObject() bool {
	self.Log.Info("DeleteReturnsDeletedObject")
	return false
}

// ([rest.CollectionDeleter] interface)
// ([rest.StandardStorage] interface)
func (self *Store) DeleteCollection(context contextpkg.Context, deleteValidation rest.ValidateObjectFunc, options *meta.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	// Note: This verb cannot be called via kubectl; test with other clients

	if options == nil {
		self.Log.Info("DeleteCollection")
	} else {
		self.Log.Infof("DeleteCollection: options=%+v", *options)
	}

	if listOptions == nil {
		listOptions = new(metainternalversion.ListOptions)
	} else {
		listOptions = listOptions.DeepCopy()
	}

	var deletedOjects []runtime.Object
	var deletedObjectsLock sync.Mutex

	deleter := util.NewParallelExecutor[runtime.Object](ParallelBufferSize, func(object runtime.Object) error {
		if accessor, err := metabase.Accessor(object); err == nil {
			name := accessor.GetName()
			self.Log.Infof("DeleteCollection: deleting %q", name)
			if _, _, err := self.Delete(context, name, deleteValidation, options); err == nil {
				deletedObjectsLock.Lock()
				deletedOjects = append(deletedOjects, object)
				deletedObjectsLock.Unlock()
			} else if apierrors.IsNotFound(err) {
				self.Log.Infof("listed item has already been deleted during DeleteCollection: %s", name)
			} else {
				return err
			}
		} else {
			return err
		}
		return nil
	})

	deleter.PanicAsError = "DeleteCollection"
	deleter.Start(ParallelWorkers)

	var errs []error
	for {
		if list, err := self.List(context, listOptions); err == nil {
			if objects, err := metabase.ExtractList(list); err == nil {
				for _, object := range objects {
					deleter.Queue(object)
				}
			} else {
				errs = append(errs, err)
			}

			if list_, err := metabase.ListAccessor(list); err == nil {
				if listOptions.Continue = list_.GetContinue(); listOptions.Continue == "" {
					break
				}
			} else {
				break
			}
		} else {
			errs = append(errs, err)
		}
	}

	errs = append(deleter.Wait(), errs...)

	list := self.NewList()
	metabase.SetList(list, deletedOjects)
	return list, errors.Join(errs...)
}

// ([rest.Creater] interface)
// ([rest.CreaterUpdater] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Create(context contextpkg.Context, object runtime.Object, createValidation rest.ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error) {
	// See: https://github.com/kubernetes/apiserver/blob/bd6de43ed55ef3094738331a1264554be65c22c9/pkg/registry/generic/registry/store.go#L399

	if options == nil {
		self.Log.Info("Creater.Create")
	} else {
		self.Log.Infof("Creater.Create: options=%+v", *options)
	}

	objectMeta, err := metabase.Accessor(object)
	if err != nil {
		return nil, apierrors.NewBadRequest(err.Error())
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

	if (options != nil) && (len(options.DryRun) > 0) {
		return nil, nil
	}

	if object, err = self.CreateFunc(context, self, object); err == nil {
		return object, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, apierrors.NewBadRequest(err.Error())
	} else {
		return nil, apierrors.NewInternalError(err)
	}
}

// ([rest.NamedCreater] interface)
func (self *Store) _Create(context contextpkg.Context, name string, object runtime.Object, createValidation rest.ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error) {
	if options == nil {
		self.Log.Infof("NamedCreater.Create: name=%s", name)
	} else {
		self.Log.Infof("NamedCreater.Create: name=%s options=%+v", name, *options)
	}
	return self.New(), nil
}

// ([rest.SubresourceObjectMetaPreserver] interface)
func (self *Store) PreserveRequestObjectMetaSystemFieldsOnSubresourceCreate() bool {
	self.Log.Info("PreserveRequestObjectMetaSystemFieldsOnSubresourceCreate")
	return false
}

// ([rest.Updater] interface)
// ([rest.CreaterUpdater] interface)
// ([rest.Patcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) Update(context contextpkg.Context, name string, objectInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	if options == nil {
		self.Log.Infof("Update: name=%s forceAllowCreate=%t", name, forceAllowCreate)
	} else {
		self.Log.Infof("Update: name=%s forceAllowCreate=%t options=%+v", name, forceAllowCreate, *options)
	}

	currentObject, err := self.Get(context, name, nil)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Note: We are purposefully ignoring forceAllowCreate
			if self.AllowCreateOnUpdate() {
				self.Log.Infof("Update: will create name=%s", name)
				currentObject = nil // just making sure
			} else {
				return nil, false, err
			}
		} else {
			return nil, false, err
		}
	}

	// Note: Running "kubectl apply --server-side" works as expected. However, without the
	// "--server-side" flag it seems that kubectl sends the updated object using the special
	// "__internal" API version, so that objectInfo.UpdatedObject() would understandably fail
	// to find the type. Our workaround is to call krm.AddInternalToScheme() on the API
	// server's Scheme (see scheme.go). Is this is a bug in kubectl? In apiserver? If not,
	// why this odd bevavior?

	updatedOrNewObject, err := objectInfo.UpdatedObject(context, currentObject)
	if err != nil {
		return nil, false, apierrors.NewInternalError(err)
	}

	if (options != nil) && (len(options.DryRun) > 0) {
		return nil, false, nil
	}

	var updateOrCreate func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error)
	if currentObject != nil {
		updateOrCreate = self.UpdateFunc
	}
	if updateOrCreate == nil {
		updateOrCreate = self.CreateFunc
	}

	if updatedOrNewObject, err = updateOrCreate(context, self, updatedOrNewObject); err == nil {
		return updatedOrNewObject, currentObject == nil, nil
	} else if backendpkg.IsBadArgumentError(err) {
		return nil, false, apierrors.NewBadRequest(err.Error())
	} else if backendpkg.IsNotFoundError(err) {
		return nil, false, apierrors.NewNotFound(self.groupResource, name)
	} else if backendpkg.IsNotDoneError(err) {
		return nil, false, apierrors.NewConflict(self.groupResource, name, err)
	} else if backendpkg.IsBusyError(err) {
		return nil, false, apierrors.NewResourceExpired(err.Error())
	} else {
		return nil, false, apierrors.NewInternalError(err)
	}
}

// ([rest.Watcher] interface)
// ([rest.StandardStorage] interface)
func (self *Store) _Watch(context contextpkg.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	if options == nil {
		self.Log.Infof("Watch")
	} else {
		self.Log.Infof("Watch: options=%+v", *options)
	}
	// TODO: supporting this would require an event model on the backend...
	return nil, nil
}

// ([rest.Redirector] interface)
func (self *Store) _ResourceLocation(context contextpkg.Context, id string) (*url.URL, http.RoundTripper, error) {
	self.Log.Infof("ResourceLocation: id=%s", id)
	return nil, nil, nil
}

// ([rest.Responder] interface)
func (self *Store) _Object(statusCode int, object runtime.Object) {
	self.Log.Infof("Object: statusCode=%d", statusCode)
}

// ([rest.Responder] interface)
func (self *Store) _Error(err error) {
	self.Log.Info("Error")
}

// ([rest.Connecter] interface)
func (self *Store) _Connect(context contextpkg.Context, id string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	self.Log.Infof("Connect: id=%s", id)
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
	self.Log.Infof("InputStream: apiVersion=%s acceptHeader=%s", apiVersion, acceptHeader)
	return nil, false, "", nil
}

// ([rest.StorageMetadata] interface)
func (self *Store) _ProducesMIMETypes(verb string) []string {
	self.Log.Infof("ProducesMIMETypes: verb=%s", verb)
	return nil
}

// ([rest.StorageMetadata] interface)
func (self *Store) _ProducesObject(verb string) any {
	self.Log.Infof("ProducesObject: verb=%s", verb)
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
func (self *Store) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
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
	self.Log.Infof("Recognizes: gvk=%s", gvk)
	return (gvk.Group == krmgroup.GroupName) && (gvk.Version == krm.Version) && (gvk.Kind == self.TypeKind)
}

// ([names.NameGenerator] interface)
// ([rest.RESTCreateStrategy] interface)
// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) GenerateName(base string) string {
	self.Log.Infof("GenerateName: base=%s", base)
	return base + backendpkg.NewID()
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) AllowCreateOnUpdate() bool {
	self.Log.Info("AllowCreateOnUpdate")
	return self.CanCreateOnUpdate
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) PrepareForUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) {
	self.Log.Info("PrepareForUpdate")
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) ValidateUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) field.ErrorList {
	self.Log.Info("ValidateUpdate")
	return nil
}

// ([rest.RESTUpdateStrategy] interface)
// ([rest.RESTCreateUpdateStrategy] interface)
// ([rest.CreateUpdateResetFieldsStrategy] interface)
// ([rest.UpdateResetFieldsStrategy] interface)
func (self *Store) WarningsOnUpdate(context contextpkg.Context, object runtime.Object, oldObject runtime.Object) []string {
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
func (self *Store) AllowUnconditionalUpdate() bool {
	self.Log.Info("AllowUnconditionalUpdate")
	return true
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
