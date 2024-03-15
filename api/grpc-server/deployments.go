package server

import (
	contextpkg "context"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ([api.APIServer] interface)
func (self *Server) CreateDeployment(context contextpkg.Context, createDeployment *api.CreateDeployment) (*api.CreateDeploymentResponse, error) {
	self.Log.Infof("createDeployment: %+v", createDeployment)

	deployment, err := backend.NewDeploymentFromBytes(createDeployment.ParentDeploymentId, createDeployment.TemplateId, createDeployment.SiteId, createDeployment.MergeMetadata, createDeployment.Prepared, createDeployment.Approved, createDeployment.MergePackageFormat, createDeployment.MergePackage)
	if err != nil {
		return new(api.CreateDeploymentResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	if err := self.Backend.CreateDeployment(context, deployment); err == nil {
		return &api.CreateDeploymentResponse{Created: true, DeploymentId: deployment.DeploymentID}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.CreateDeploymentResponse{Created: false, NotCreatedReason: err.Error()}, nil
	} else {
		return new(api.CreateDeploymentResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeleteDeployment(context contextpkg.Context, deploymentId *api.DeploymentID) (*api.DeleteResponse, error) {
	self.Log.Infof("deleteDeployment: %+v", deploymentId)

	if err := self.Backend.DeleteDeployment(context, deploymentId.DeploymentId); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetDeployment(context contextpkg.Context, getDeployment *api.GetDeployment) (*api.Deployment, error) {
	self.Log.Infof("getDeployment: %+v", getDeployment)

	if deployment, err := self.Backend.GetDeployment(context, getDeployment.DeploymentId); err == nil {
		packageFormat := getDeployment.PreferredPackageFormat
		if packageFormat == "" {
			packageFormat = self.DefaultPackageFormat
		}
		if pakcage_, err := deployment.EncodePackage(packageFormat); err == nil {
			return &api.Deployment{
				DeploymentId:       deployment.DeploymentID,
				ParentDeploymentId: deployment.ParentDeploymentID,
				TemplateId:         deployment.TemplateID,
				SiteId:             deployment.SiteID,
				Created:            timestamppb.New(deployment.Created),
				Updated:            timestamppb.New(deployment.Updated),
				Prepared:           deployment.Prepared,
				Approved:           deployment.Approved,
				PackageFormat:      packageFormat,
				Package:            pakcage_,
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
	self.Log.Infof("listDeployments: %+v", listDeployments)

	if deploymentInfoResults, err := self.Backend.ListDeployments(server.Context(), backend.SelectDeployments{
		ParentDeploymentID:       listDeployments.Select.ParentDeploymentId,
		MetadataPatterns:         listDeployments.Select.MetadataPatterns,
		TemplateIDPatterns:       listDeployments.Select.TemplateIdPatterns,
		TemplateMetadataPatterns: listDeployments.Select.TemplateMetadataPatterns,
		SiteIDPatterns:           listDeployments.Select.SiteIdPatterns,
		SiteMetadataPatterns:     listDeployments.Select.SiteMetadataPatterns,
		Prepared:                 listDeployments.Select.Prepared,
		Approved:                 listDeployments.Select.Approved,
	}, backend.Window{
		Offset:   uint(listDeployments.Window.Offset),
		MaxCount: int(listDeployments.Window.MaxCount),
	}); err == nil {
		if err := util.IterateResults(deploymentInfoResults, func(deploymentInfo backend.DeploymentInfo) error {
			return server.Send(&api.ListedDeployment{
				DeploymentId:       deploymentInfo.DeploymentID,
				ParentDeploymentId: deploymentInfo.ParentDeploymentID,
				TemplateId:         deploymentInfo.TemplateID,
				SiteId:             deploymentInfo.SiteID,
				Metadata:           deploymentInfo.Metadata,
				Created:            timestamppb.New(deploymentInfo.Created),
				Updated:            timestamppb.New(deploymentInfo.Updated),
				Prepared:           deploymentInfo.Prepared,
				Approved:           deploymentInfo.Approved,
			})
		}); err != nil {
			return ToGRPCError(err)
		}
	} else {
		return ToGRPCError(err)
	}

	return nil
}

// ([api.APIServer] interface)
func (self *Server) PurgeDeployments(context contextpkg.Context, selectDeployments *api.SelectDeployments) (*api.DeleteResponse, error) {
	self.Log.Infof("purgeDeployments: %+v", selectDeployments)

	if err := self.Backend.PurgeDeployments(context, backend.SelectDeployments{
		ParentDeploymentID:       selectDeployments.ParentDeploymentId,
		MetadataPatterns:         selectDeployments.MetadataPatterns,
		TemplateIDPatterns:       selectDeployments.TemplateIdPatterns,
		TemplateMetadataPatterns: selectDeployments.TemplateMetadataPatterns,
		SiteIDPatterns:           selectDeployments.SiteIdPatterns,
		SiteMetadataPatterns:     selectDeployments.SiteMetadataPatterns,
		Prepared:                 selectDeployments.Prepared,
		Approved:                 selectDeployments.Approved,
	}); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) StartDeploymentModification(context contextpkg.Context, startDeploymentModification *api.StartDeploymentModification) (*api.StartDeploymentModificationResponse, error) {
	self.Log.Infof("startDeploymentModification: %+v", startDeploymentModification)

	if modificationToken, deployment, err := self.Backend.StartDeploymentModification(context, startDeploymentModification.DeploymentId); err == nil {
		packageFormat := startDeploymentModification.PreferredPackageFormat
		if packageFormat == "" {
			packageFormat = self.DefaultPackageFormat
		}
		if package_, err := deployment.EncodePackage(packageFormat); err == nil {
			return &api.StartDeploymentModificationResponse{
				Started:           true,
				ModificationToken: modificationToken,
				PackageFormat:     packageFormat,
				Package:           package_,
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
	self.Log.Infof("endDeploymentModification: %+v", endDeploymentModification)

	package_, err := tkoutil.DecodePackage(endDeploymentModification.PackageFormat, endDeploymentModification.Package)
	if err != nil {
		return new(api.EndDeploymentModificationResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	if deploymentId, err := self.Backend.EndDeploymentModification(context, endDeploymentModification.ModificationToken, package_, nil); err == nil {
		return &api.EndDeploymentModificationResponse{Modified: true, DeploymentId: deploymentId}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.EndDeploymentModificationResponse{Modified: false, NotModifiedReason: err.Error()}, nil
	} else {
		return new(api.EndDeploymentModificationResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) CancelDeploymentModification(context contextpkg.Context, cancelDeploymentModification *api.CancelDeploymentModification) (*api.CancelDeploymentModificationResponse, error) {
	self.Log.Infof("cancelDeploymentModification: %+v", cancelDeploymentModification)

	if err := self.Backend.CancelDeploymentModification(context, cancelDeploymentModification.ModificationToken); err == nil {
		return &api.CancelDeploymentModificationResponse{Cancelled: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.CancelDeploymentModificationResponse{Cancelled: false, NotCancelledReason: err.Error()}, nil
	} else {
		return new(api.CancelDeploymentModificationResponse), ToGRPCError(err)
	}
}
