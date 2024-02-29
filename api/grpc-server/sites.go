package server

import (
	contextpkg "context"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ([api.APIServer] interface)
func (self *Server) RegisterSite(context contextpkg.Context, site *api.Site) (*api.RegisterResponse, error) {
	self.Log.Infof("registerSite: %+v", site)

	site_, err := backend.NewSiteFromBytes(site.SiteId, site.TemplateId, site.Metadata, site.ResourcesFormat, site.Resources)
	if err != nil {
		return new(api.RegisterResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	site_.UpdateFromResources()

	if err := self.Backend.SetSite(context, site_); err == nil {
		return &api.RegisterResponse{Registered: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.RegisterResponse{Registered: false, NotRegisteredReason: err.Error()}, nil
	} else {
		return new(api.RegisterResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) DeleteSite(context contextpkg.Context, siteId *api.SiteID) (*api.DeleteResponse, error) {
	self.Log.Infof("deleteSite: %+v", siteId)

	if err := self.Backend.DeleteSite(context, siteId.SiteId); err == nil {
		return &api.DeleteResponse{Deleted: true}, nil
	} else if backend.IsNotDoneError(err) {
		return &api.DeleteResponse{Deleted: false, NotDeletedReason: err.Error()}, nil
	} else {
		return new(api.DeleteResponse), ToGRPCError(err)
	}
}

// ([api.APIServer] interface)
func (self *Server) GetSite(context contextpkg.Context, getSite *api.GetSite) (*api.Site, error) {
	self.Log.Infof("getSite: %+v", getSite)

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
	self.Log.Infof("listSites: %+v", listSites)

	if siteInfoResults, err := self.Backend.ListSites(server.Context(), backend.ListSites{
		Offset:             uint(listSites.Offset),
		MaxCount:           uint(listSites.MaxCount),
		SiteIDPatterns:     listSites.SiteIdPatterns,
		TemplateIDPatterns: listSites.TemplateIdPatterns,
		MetadataPatterns:   listSites.MetadataPatterns,
	}); err == nil {
		if err := util.IterateResults(siteInfoResults, func(siteInfo backend.SiteInfo) error {
			return server.Send(&api.ListedSite{
				SiteId:        siteInfo.SiteID,
				TemplateId:    siteInfo.TemplateID,
				Metadata:      siteInfo.Metadata,
				DeploymentIds: siteInfo.DeploymentIDs,
			})
		}); err != nil {
			return ToGRPCError(err)
		}
	} else {
		return ToGRPCError(err)
	}

	return nil
}
