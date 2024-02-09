package server

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ([api.APIServer] interface)
func (self *Server) CreateDeployment(context contextpkg.Context, createDeployment *api.CreateDeployment) (*api.CreateDeploymentResponse, error) {
	self.Log.Infof("createDeployment: %s", createDeployment)

	deployment, err := backend.NewDeploymentFromBytes(createDeployment.TemplateId, createDeployment.ParentDeploymentId, createDeployment.SiteId, createDeployment.MergeMetadata, createDeployment.Prepared, createDeployment.Approved, createDeployment.MergeResourcesFormat, createDeployment.MergeResources)
	if err != nil {
		return new(api.CreateDeploymentResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	if err := self.Backend.SetDeployment(context, deployment); err == nil {
		return &api.CreateDeploymentResponse{Created: true, DeploymentId: deployment.DeploymentID}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.CreateDeploymentResponse{Created: false, NotCreatedReason: err.Error()}, nil
	} else {
		return new(api.CreateDeploymentResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeleteDeployment(context contextpkg.Context, deleteDeployment *api.DeleteDeployment) (*api.DeleteResponse, error) {
	self.Log.Infof("deleteDeployment: %s", deleteDeployment)

	if err := self.Backend.DeleteDeployment(context, deleteDeployment.DeploymentId); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetDeployment(context contextpkg.Context, getDeployment *api.GetDeployment) (*api.Deployment, error) {
	self.Log.Infof("getDeployment: %s", getDeployment)

	if deployment, err := self.Backend.GetDeployment(context, getDeployment.DeploymentId); err == nil {
		resourcesFormat := getDeployment.PreferredResourcesFormat
		if resourcesFormat == "" {
			resourcesFormat = self.DefaultResourcesFormat
		}
		if resources, err := deployment.EncodeResources(resourcesFormat); err == nil {
			return &api.Deployment{
				DeploymentId:       deployment.DeploymentID,
				ParentDeploymentId: deployment.ParentDeploymentID,
				TemplateId:         deployment.TemplateID,
				SiteId:             deployment.SiteID,
				Prepared:           deployment.Prepared,
				Approved:           deployment.Approved,
				ResourcesFormat:    resourcesFormat,
				Resources:          resources,
			}, nil
		} else {
			return new(api.Deployment), ToGRPCError(err)
		}
	} else {
		return new(api.Deployment), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) ListDeployments(listDeployments *api.ListDeployments, server api.API_ListDeploymentsServer) error {
	self.Log.Infof("listDeployments: %s", listDeployments)

	if deploymentInfos, err := self.Backend.ListDeployments(server.Context(), backend.ListDeployments{
		ParentDeploymentID:       listDeployments.ParentDeploymentId,
		MetadataPatterns:         listDeployments.MetadataPatterns,
		TemplateIDPatterns:       listDeployments.TemplateIdPatterns,
		TemplateMetadataPatterns: listDeployments.TemplateMetadataPatterns,
		SiteIDPatterns:           listDeployments.SiteIdPatterns,
		SiteMetadataPatterns:     listDeployments.SiteMetadataPatterns,
		Prepared:                 listDeployments.Prepared,
		Approved:                 listDeployments.Approved,
	}); err == nil {
		for _, deploymentInfo := range deploymentInfos {
			if err := server.Send(&api.ListDeploymentsResponse{
				DeploymentId:       deploymentInfo.DeploymentID,
				ParentDeploymentId: deploymentInfo.ParentDeploymentID,
				TemplateId:         deploymentInfo.TemplateID,
				SiteId:             deploymentInfo.SiteID,
				Metadata:           deploymentInfo.Metadata,
				Prepared:           deploymentInfo.Prepared,
				Approved:           deploymentInfo.Approved,
			}); err != nil {
				return err
			}
		}
	} else {
		return ToGRPCError(err)
	}

	return nil
}

// ([api.APIServer] interface)
func (self *Server) StartDeploymentModification(context contextpkg.Context, startDeploymentModification *api.StartDeploymentModification) (*api.StartDeploymentModificationResponse, error) {
	self.Log.Infof("startDeploymentModification: %s", startDeploymentModification)

	if modificationToken, deployment, err := self.Backend.StartDeploymentModification(context, startDeploymentModification.DeploymentId); err == nil {
		resourcesFormat := startDeploymentModification.PreferredResourcesFormat
		if resourcesFormat == "" {
			resourcesFormat = self.DefaultResourcesFormat
		}
		if resources, err := deployment.EncodeResources(resourcesFormat); err == nil {
			return &api.StartDeploymentModificationResponse{
				Started:           true,
				ModificationToken: modificationToken,
				ResourcesFormat:   resourcesFormat,
				Resources:         resources,
			}, nil
		} else {
			return new(api.StartDeploymentModificationResponse), ToGRPCError(err)
		}
	} else if backend.IsNotDoneError(err) {
		return &api.StartDeploymentModificationResponse{Started: false, NotStartedReason: err.Error()}, nil
	} else {
		return new(api.StartDeploymentModificationResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) EndDeploymentModification(context contextpkg.Context, endDeploymentModification *api.EndDeploymentModification) (*api.EndDeploymentModificationResponse, error) {
	self.Log.Infof("endDeploymentModification: %s", endDeploymentModification)

	resources, err := util.DecodeResources(endDeploymentModification.ResourcesFormat, endDeploymentModification.Resources)
	if err != nil {
		return new(api.EndDeploymentModificationResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	if deploymentId, err := self.Backend.EndDeploymentModification(context, endDeploymentModification.ModificationToken, resources); err == nil {
		return &api.EndDeploymentModificationResponse{Modified: true, DeploymentId: deploymentId}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.EndDeploymentModificationResponse{Modified: false, NotModifiedReason: err.Error()}, nil
	} else {
		return new(api.EndDeploymentModificationResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) CancelDeploymentModification(context contextpkg.Context, cancelDeploymentModification *api.CancelDeploymentModification) (*api.CancelDeploymentModificationResponse, error) {
	self.Log.Infof("cancelDeploymentModification: %s", cancelDeploymentModification)

	if err := self.Backend.CancelDeploymentModification(context, cancelDeploymentModification.ModificationToken); err == nil {
		return &api.CancelDeploymentModificationResponse{Cancelled: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.CancelDeploymentModificationResponse{Cancelled: false, NotCancelledReason: err.Error()}, nil
	} else {
		return new(api.CancelDeploymentModificationResponse), ToGRPCError(err)
	}
}
