package server

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	api "github.com/nephio-experimental/tko/api/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ([api.APIServer] interface)
func (self *Server) RegisterSite(context contextpkg.Context, site *api.Site) (*api.RegisterResponse, error) {
	self.Log.Infof("registerSite: %s", site)

	site_, err := backend.NewSiteFromBytes(site.SiteId, site.TemplateId, site.Metadata, site.ResourcesFormat, site.Resources)
	if err != nil {
		return new(api.RegisterResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	if err := self.Backend.SetSite(context, site_); err == nil {
		return &api.RegisterResponse{Registered: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.RegisterResponse{Registered: false, NotRegisteredReason: err.Error()}, nil
	} else {
		return new(api.RegisterResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeleteSite(context contextpkg.Context, deleteSite *api.DeleteSite) (*api.DeleteResponse, error) {
	self.Log.Infof("deleteSite: %s", deleteSite)

	if err := self.Backend.DeleteSite(context, deleteSite.SiteId); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetSite(context contextpkg.Context, getSite *api.GetSite) (*api.Site, error) {
	self.Log.Infof("getSite: %s", getSite)

	if site, err := self.Backend.GetSite(context, getSite.SiteId); err == nil {
		resourcesFormat := getSite.PreferredResourcesFormat
		if resourcesFormat == "" {
			resourcesFormat = self.DefaultResourcesFormat
		}
		if resources, err := site.EncodeResources(resourcesFormat); err == nil {
			return &api.Site{
				SiteId:          site.SiteID,
				TemplateId:      site.TemplateID,
				Metadata:        site.Metadata,
				ResourcesFormat: resourcesFormat,
				Resources:       resources,
				DeploymentIds:   site.DeploymentIDs,
			}, nil
		} else {
			return new(api.Site), ToGRPCError(err)
		}
	} else {
		return new(api.Site), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) ListSites(listSites *api.ListSites, server api.API_ListSitesServer) error {
	self.Log.Infof("listSites: %s", listSites)

	if siteInfoStream, err := self.Backend.ListSites(server.Context(), backend.ListSites{
		SiteIDPatterns:     listSites.SiteIdPatterns,
		TemplateIDPatterns: listSites.TemplateIdPatterns,
		MetadataPatterns:   listSites.MetadataPatterns,
	}); err == nil {
		for {
			if siteInfo, ok := siteInfoStream.Next(); ok {
				if err := server.Send(&api.ListSitesResponse{
					SiteId:        siteInfo.SiteID,
					TemplateId:    siteInfo.TemplateID,
					Metadata:      siteInfo.Metadata,
					DeploymentIds: siteInfo.DeploymentIDs,
				}); err != nil {
					return err
				}
			} else {
				break
			}
		}
	} else {
		return ToGRPCError(err)
	}

	return nil
}
