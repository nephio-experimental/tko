package server

import (
	contextpkg "context"
	"fmt"
	"time"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
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

func NewDeploymentStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:          "Deployment",
		TypeListKind:      "DeploymentList",
		TypeSingular:      "deployment",
		TypePlural:        "deployments",
		CanCreateOnUpdate: false,
		ObjectTyper:       Scheme,

		NewObjectFunc: func() runtime.Object {
			return new(krm.Deployment)
		},

		NewListObjectFunc: func() runtime.Object {
			return new(krm.DeploymentList)
		},

		GetFieldsFunc: func(object runtime.Object) (fields.Set, error) {
			if krmDeployment, ok := object.(*krm.Deployment); ok {
				fields := fields.Set{
					"metadata.name": krmDeployment.Name,
				}
				if krmDeployment.Spec.TemplateId != nil {
					fields["spec.templateId"] = *krmDeployment.Spec.TemplateId
				}
				return fields, nil
			} else {
				return nil, fmt.Errorf("not a deployment: %T", object)
			}
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if deployment, err := DeploymentFromKRM(object); err == nil {
				if err := store.Backend.CreateDeployment(context, deployment); err == nil {
					return object, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		UpdateFunc: func(context contextpkg.Context, store *Store, updatedObject runtime.Object) (runtime.Object, error) {
			if updatedDeployment, err := DeploymentFromKRM(updatedObject); err == nil {
				if modificationToken, deployment, err := store.Backend.StartDeploymentModification(context, updatedDeployment.DeploymentID); err == nil {
					if ResourceVersionsEqual(deployment.Updated, updatedDeployment.Updated) {
						if _, err := store.Backend.EndDeploymentModification(context, modificationToken, updatedDeployment.Package, nil); err == nil {
							return updatedObject, nil
						} else {
							return nil, err
						}
					} else if err := store.Backend.CancelDeploymentModification(context, modificationToken); err == nil {
						return nil, backendpkg.NewNotDoneErrorf("deployment has been modified before the proposed update: %s, was %s", deployment.Updated, updatedDeployment.Updated)
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteDeployment(context, id)
		},

		PurgeFunc: func(context contextpkg.Context, store *Store) error {
			return store.Backend.PurgeDeployments(context, backendpkg.SelectDeployments{})
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if deployment, err := store.Backend.GetDeployment(context, id); err == nil {
				if krmDeployment, err := DeploymentToKRM(deployment); err == nil {
					return krmDeployment, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, options *metainternalversion.ListOptions, offset uint, maxCount uint) (runtime.Object, error) {
			var krmDeploymentList krm.DeploymentList

			var metadataPatterns map[string]string
			var err error
			if metadataPatterns, err = ToMetadataPatterns(options); err != nil {
				return nil, err
			}
			selectionPredicate := store.NewSelectionPredicate(options, false)

			if results, err := store.Backend.ListDeployments(context, backendpkg.SelectDeployments{MetadataPatterns: metadataPatterns}, backendpkg.Window{Offset: offset, MaxCount: int(maxCount)}); err == nil {
				if err := util.IterateResults(results, func(deploymentInfo backendpkg.DeploymentInfo) error {
					if krmDeployment, err := DeploymentInfoToKRM(&deploymentInfo); err == nil {
						if ok, err := selectionPredicate.Matches(krmDeployment); err == nil {
							if ok {
								krmDeploymentList.Items = append(krmDeploymentList.Items, *krmDeployment)
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

			krmDeploymentList.APIVersion = APIVersion
			krmDeploymentList.Kind = "DeploymentList"
			return &krmDeploymentList, nil
		},

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, withHeaders bool, withObject bool) (*meta.Table, error) {
			table := new(meta.Table)

			krmDeployments, err := ToDeploymentsKRM(object)
			if err != nil {
				return nil, err
			}

			if withHeaders {
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name"},
					{Name: "ParentDeploymentID", Type: "string"},
					{Name: "TemplateID", Type: "string"},
					{Name: "SiteID", Type: "string"},
					{Name: "Prepared", Type: "boolean"},
					{Name: "Approved", Type: "boolean"},
					{Name: "Created", Type: "string", Format: "date-time"},
					{Name: "Updated", Type: "string", Format: "date-time"},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmDeployments))
			for index, krmDeployment := range krmDeployments {
				var updated time.Time
				var err error
				if updated, err = FromResourceVersion(krmDeployment.ResourceVersion); err != nil {
					return nil, err
				}

				row := meta.TableRow{
					Cells: []any{
						krmDeployment.Name,
						krmDeployment.Spec.ParentDeploymentId,
						krmDeployment.Spec.TemplateId,
						krmDeployment.Spec.SiteId,
						krmDeployment.Status.Prepared,
						krmDeployment.Status.Approved,
						// Note: kubectl will always display these in UTC, even if we change the timezone here
						krmDeployment.CreationTimestamp,
						updated,
					},
				}
				if withObject {
					row.Object = runtime.RawExtension{Object: &krmDeployment}
				}
				table.Rows[index] = row
			}

			return table, nil
		},
	}

	store.Init()
	return &store
}

func ToDeploymentsKRM(object runtime.Object) ([]krm.Deployment, error) {
	switch object_ := object.(type) {
	case *krm.DeploymentList:
		return object_.Items, nil
	case *krm.Deployment:
		return []krm.Deployment{*object_}, nil
	default:
		return nil, backendpkg.NewBadArgumentErrorf("unsupported type: %T", object)
	}
}

func DeploymentInfoToKRM(deploymentInfo *backendpkg.DeploymentInfo) (*krm.Deployment, error) {
	name, err := tkoutil.ToKubernetesName(deploymentInfo.DeploymentID)
	if err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var krmDeployment krm.Deployment
	krmDeployment.APIVersion = APIVersion
	krmDeployment.Kind = "Deployment"
	krmDeployment.Name = name
	krmDeployment.UID = types.UID("tko|deployment|" + deploymentInfo.DeploymentID)
	krmDeployment.CreationTimestamp = meta.NewTime(deploymentInfo.Created)
	krmDeployment.ResourceVersion = ToResourceVersion(deploymentInfo.Updated)
	krmDeployment.Labels, _ = tkoutil.ToKubernetesNames(deploymentInfo.Metadata)

	deploymentId := deploymentInfo.DeploymentID
	krmDeployment.Spec.DeploymentId = &deploymentId
	if parentDeploymentId := deploymentInfo.ParentDeploymentID; parentDeploymentId != "" {
		krmDeployment.Spec.ParentDeploymentId = &parentDeploymentId
	}
	if templateId := deploymentInfo.TemplateID; templateId != "" {
		krmDeployment.Spec.TemplateId = &templateId
	}
	if siteId := deploymentInfo.SiteID; siteId != "" {
		krmDeployment.Spec.SiteId = &siteId
	}
	prepared := deploymentInfo.Prepared
	krmDeployment.Status.Prepared = &prepared
	approved := deploymentInfo.Approved
	krmDeployment.Status.Approved = &approved

	return &krmDeployment, nil
}

func DeploymentToKRM(deployment *backendpkg.Deployment) (*krm.Deployment, error) {
	if krmDeployment, err := DeploymentInfoToKRM(&deployment.DeploymentInfo); err == nil {
		krmDeployment.Spec.Package = PackageToKRM(deployment.Package)
		return krmDeployment, nil
	} else {
		return nil, err
	}
}

func DeploymentFromKRM(object runtime.Object) (*backendpkg.Deployment, error) {
	var krmDeployment *krm.Deployment
	var ok bool
	if krmDeployment, ok = object.(*krm.Deployment); !ok {
		return nil, backendpkg.NewBadArgumentErrorf("not a Deployment: %T", object)
	}

	var deploymentId string
	var err error
	if deploymentId, err = tkoutil.FromKubernetesName(krmDeployment.Name); err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var updated time.Time
	if updated, err = FromResourceVersion(krmDeployment.ResourceVersion); err != nil {
		return nil, err
	}

	deployment := backendpkg.Deployment{
		DeploymentInfo: backendpkg.DeploymentInfo{
			DeploymentID: deploymentId,
			Created:      krmDeployment.CreationTimestamp.Time,
			Updated:      updated,
		},
	}

	deployment.Metadata, _ = tkoutil.FromKubernetesNames(krmDeployment.Labels)
	deployment.Package = PackageFromKRM(krmDeployment.Spec.Package)

	if krmDeployment.Spec.ParentDeploymentId != nil {
		deployment.ParentDeploymentID = *krmDeployment.Spec.ParentDeploymentId
	}

	if krmDeployment.Spec.TemplateId != nil {
		deployment.TemplateID = *krmDeployment.Spec.TemplateId
	}

	if krmDeployment.Spec.SiteId != nil {
		deployment.SiteID = *krmDeployment.Spec.SiteId
	}

	if krmDeployment.Status.Prepared != nil {
		deployment.Prepared = *krmDeployment.Status.Prepared
	}

	if krmDeployment.Status.Approved != nil {
		deployment.Approved = *krmDeployment.Status.Approved
	}

	return &deployment, nil
}
