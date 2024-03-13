package server

import (
	contextpkg "context"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ([api.APIServer] interface)
func (self *Server) RegisterTemplate(context contextpkg.Context, template *api.Template) (*api.RegisterResponse, error) {
	self.Log.Infof("registerTemplate: %+v", template)

	template_, err := backend.NewTemplateFromBytes(template.TemplateId, template.Metadata, template.PackageFormat, template.Package)
	if err != nil {
		return new(api.RegisterResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	template_.UpdateFromPackage()

	if err := self.Backend.SetTemplate(context, template_); err == nil {
		return &api.RegisterResponse{Registered: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.RegisterResponse{Registered: false, NotRegisteredReason: err.Error()}, nil
	} else {
		return new(api.RegisterResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeleteTemplate(context contextpkg.Context, templateId *api.TemplateID) (*api.DeleteResponse, error) {
	self.Log.Infof("deleteTemplate: %+v", templateId)

	if err := self.Backend.DeleteTemplate(context, templateId.TemplateId); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetTemplate(context contextpkg.Context, getTemplate *api.GetTemplate) (*api.Template, error) {
	self.Log.Infof("getTemplate: %+v", getTemplate)

	if template, err := self.Backend.GetTemplate(context, getTemplate.TemplateId); err == nil {
		packageFormat := getTemplate.PreferredPackageFormat
		if packageFormat == "" {
			packageFormat = self.DefaultPackageFormat
		}
		if package_, err := template.EncodePackage(packageFormat); err == nil {
			return &api.Template{
				TemplateId:    template.TemplateID,
				Metadata:      template.Metadata,
				Updated:       timestamppb.New(template.Updated),
				PackageFormat: packageFormat,
				Package:       package_,
				DeploymentIds: template.DeploymentIDs,
			}, nil
		} else {
			return new(api.Template), ToGRPCError(err)
		}
	} else {
		return new(api.Template), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) ListTemplates(listTemplates *api.ListTemplates, server api.API_ListTemplatesServer) error {
	self.Log.Infof("listTemplates: %+v", listTemplates)

	if templateInfoResults, err := self.Backend.ListTemplates(server.Context(), backend.SelectTemplates{
		TemplateIDPatterns: listTemplates.Select.TemplateIdPatterns,
		MetadataPatterns:   listTemplates.Select.MetadataPatterns,
	}, backend.Window{
		Offset:   uint(listTemplates.Window.Offset),
		MaxCount: uint(listTemplates.Window.MaxCount),
	}); err == nil {
		if err := util.IterateResults(templateInfoResults, func(templateInfo backend.TemplateInfo) error {
			return server.Send(&api.ListedTemplate{
				TemplateId:    templateInfo.TemplateID,
				Metadata:      templateInfo.Metadata,
				Updated:       timestamppb.New(templateInfo.Updated),
				DeploymentIds: templateInfo.DeploymentIDs,
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
func (self *Server) PurgeTemplates(context contextpkg.Context, selectTemplates *api.SelectTemplates) (*api.DeleteResponse, error) {
	self.Log.Infof("purgeTemplates: %+v", selectTemplates)

	if err := self.Backend.PurgeTemplates(context, backend.SelectTemplates{
		TemplateIDPatterns: selectTemplates.TemplateIdPatterns,
		MetadataPatterns:   selectTemplates.MetadataPatterns,
	}); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}
