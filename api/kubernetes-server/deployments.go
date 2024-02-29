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

func NewDeploymentStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:     "Deployment",
		TypeListKind: "DeploymentList",
		TypeSingular: "deployment",
		TypePlural:   "deployments",
		ObjectTyper:  Scheme,

		NewResourceFunc: func() runtime.Object {
			return new(krm.Deployment)
		},

		NewResourceListFunc: func() runtime.Object {
			return new(krm.DeploymentList)
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if krmDeployment, ok := object.(*krm.Deployment); ok {
				if deployment, err := KRMToDeployment(krmDeployment); err == nil {
					if err := store.Backend.CreateDeployment(context, deployment); err == nil {
						return krmDeployment, nil
					} else {
						return nil, err
					}
				} else {
					return nil, backendpkg.NewBadArgumentError(err.Error())
				}
			} else {
				return nil, backendpkg.NewBadArgumentErrorf("not a Deployment: %T", object)
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteDeployment(context, id)
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if deployment, err := store.Backend.GetDeployment(context, id); err == nil {
				if krmDeployment, err := DeploymentToKRM(deployment); err == nil {
					return &krmDeployment, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, offset uint, maxCount uint) (runtime.Object, error) {
			var krmDeploymentList krm.DeploymentList
			krmDeploymentList.APIVersion = APIVersion
			krmDeploymentList.Kind = "DeploymentList"

			if results, err := store.Backend.ListDeployments(context, backendpkg.ListDeployments{Offset: offset, MaxCount: maxCount}); err == nil {
				if err := util.IterateResults(results, func(deploymentInfo backendpkg.DeploymentInfo) error {
					if krmDeployment, err := DeploymentInfoToKRM(&deploymentInfo); err == nil {
						krmDeploymentList.Items = append(krmDeploymentList.Items, krmDeployment)
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
					{Name: "DeploymentID", Type: "string"},
					{Name: "ParentDeploymentID", Type: "string"},
					{Name: "TemplateID", Type: "string"},
					{Name: "SiteID", Type: "string"},
					{Name: "Prepared", Type: "boolean"},
					{Name: "Approved", Type: "boolean"},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmDeployments))
			for index, krmDeployment := range krmDeployments {
				row := meta.TableRow{
					Cells: []any{
						krmDeployment.Name,
						krmDeployment.Spec.DeploymentId,
						krmDeployment.Spec.ParentDeploymentId,
						krmDeployment.Spec.TemplateId,
						krmDeployment.Spec.SiteId,
						krmDeployment.Status.Prepared,
						krmDeployment.Status.Approved,
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
		return nil, fmt.Errorf("unsupported type: %T", object)
	}
}

func DeploymentInfoToKRM(deploymentInfo *backendpkg.DeploymentInfo) (krm.Deployment, error) {
	name, err := tkoutil.ToKubernetesName(deploymentInfo.DeploymentID)
	if err != nil {
		return krm.Deployment{}, err
	}

	var krmDeployment krm.Deployment
	krmDeployment.APIVersion = APIVersion
	krmDeployment.Kind = "Deployment"
	krmDeployment.Name = name
	krmDeployment.UID = types.UID("tko|deployment|" + deploymentInfo.DeploymentID)

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
	krmDeployment.Spec.Metadata = deploymentInfo.Metadata
	prepared := deploymentInfo.Prepared
	krmDeployment.Status.Prepared = &prepared
	approved := deploymentInfo.Approved
	krmDeployment.Status.Approved = &approved

	return krmDeployment, nil
}

func DeploymentToKRM(deployment *backendpkg.Deployment) (krm.Deployment, error) {
	if krmDeployment, err := DeploymentInfoToKRM(&deployment.DeploymentInfo); err == nil {
		krmDeployment.Spec.Package = ResourcesToKRM(deployment.Resources)
		return krmDeployment, nil
	} else {
		return krm.Deployment{}, err
	}
}

func KRMToDeployment(krmDeployment *krm.Deployment) (*backendpkg.Deployment, error) {
	var deploymentId string
	var err error
	if deploymentId, err = tkoutil.FromKubernetesName(krmDeployment.Name); err != nil {
		return nil, err
	}

	deployment := backendpkg.Deployment{
		DeploymentInfo: backendpkg.DeploymentInfo{
			DeploymentID: deploymentId,
			Metadata:     krmDeployment.Spec.Metadata,
		},
	}

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

	deployment.Resources = ResourcesFromKRM(krmDeployment.Spec.Package)

	return &deployment, nil
}
