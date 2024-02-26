package server

import (
	contextpkg "context"
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	backendpkg "github.com/nephio-experimental/tko/backend"
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

		Kind:        "Deployment",
		ListKind:    "DeploymentList",
		Singular:    "deployment",
		Plural:      "deployments",
		ObjectTyper: Scheme,

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

		ListFunc: func(context contextpkg.Context, store *Store) (runtime.Object, error) {
			var krmDeploymentList krm.DeploymentList
			krmDeploymentList.APIVersion = APIVersion
			krmDeploymentList.Kind = "DeploymentList"

			if results, err := store.Backend.ListDeployments(context, backendpkg.ListDeployments{}); err == nil {
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

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, options *meta.TableOptions) (*meta.Table, error) {
			table := new(meta.Table)

			krmDeployments, err := ToDeploymentsKRM(object)
			if err != nil {
				return nil, err
			}

			if (options == nil) || !options.NoHeaders {
				descriptions := krm.Deployment{}.TypeMeta.SwaggerDoc()
				nameDescription, _ := descriptions["name"]
				deploymentIdDescription, _ := descriptions["deploymentId"]
				parentDeploymentIdDescription, _ := descriptions["parentDeploymentId"]
				templateIdDescription, _ := descriptions["templateId"]
				siteIdDescription, _ := descriptions["siteId"]
				preparedDescription, _ := descriptions["prepared"]
				approvedDescription, _ := descriptions["approved"]
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name", Description: nameDescription},
					{Name: "DeploymentID", Type: "string", Description: deploymentIdDescription},
					{Name: "ParentDeploymentID", Type: "string", Description: parentDeploymentIdDescription},
					{Name: "TemplateID", Type: "string", Description: templateIdDescription},
					{Name: "SiteID", Type: "string", Description: siteIdDescription},
					{Name: "Prepared", Type: "boolean", Description: preparedDescription},
					{Name: "Approved", Type: "boolean", Description: approvedDescription},
					//{Name: "Metadata", Description: descriptions["metadata"]},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmDeployments))
			for index, krmDeployment := range krmDeployments {
				row := meta.TableRow{
					Cells: []any{krmDeployment.Name, krmDeployment.Spec.DeploymentId, krmDeployment.Spec.ParentDeploymentId, krmDeployment.Spec.TemplateId, krmDeployment.Spec.SiteId, krmDeployment.Spec.Prepared, krmDeployment.Spec.Approved},
				}
				if (options == nil) || (options.IncludeObject != meta.IncludeNone) {
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
	name, err := IDToName(deploymentInfo.DeploymentID)
	if err != nil {
		return krm.Deployment{}, err
	}

	var krmDeployment krm.Deployment
	krmDeployment.APIVersion = APIVersion
	krmDeployment.Kind = "Deployment"
	krmDeployment.Name = name
	krmDeployment.UID = types.UID("tko|deployment|" + deploymentInfo.DeploymentID)

	if deploymentId := deploymentInfo.DeploymentID; deploymentId != "" {
		krmDeployment.Spec.DeploymentId = &deploymentId
	}

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
	krmDeployment.Spec.Prepared = &prepared
	approved := deploymentInfo.Approved
	krmDeployment.Spec.Approved = &approved

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
	var id string
	if krmDeployment.Spec.DeploymentId != nil {
		id = *krmDeployment.Spec.DeploymentId
	}
	if id == "" {
		var err error
		if id, err = NameToID(krmDeployment.Name); err != nil {
			return nil, err
		}
	}

	deployment := backendpkg.Deployment{
		DeploymentInfo: backendpkg.DeploymentInfo{
			DeploymentID: id,
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

	if krmDeployment.Spec.Prepared != nil {
		deployment.Prepared = *krmDeployment.Spec.Prepared
	}

	if krmDeployment.Spec.Approved != nil {
		deployment.Approved = *krmDeployment.Spec.Approved
	}

	deployment.Resources = ResourcesFromKRM(krmDeployment.Spec.Package)

	return &deployment, nil
}
