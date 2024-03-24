package server

import (
	"net/http"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListDeployments(writer http.ResponseWriter, request *http.Request) {
	// TODO: paging
	if deploymentInfoResults, err := self.Backend.ListDeployments(request.Context(), backend.SelectDeployments{}, getWindow(request)); err == nil {
		var deployments []ard.StringMap
		if err := util.IterateResults(deploymentInfoResults, func(deploymentInfo backend.DeploymentInfo) error {
			deployments = append(deployments, ard.StringMap{
				"id":               deploymentInfo.DeploymentID,
				"template":         deploymentInfo.TemplateID,
				"parent":           deploymentInfo.ParentDeploymentID,
				"site":             deploymentInfo.SiteID,
				"metadata":         deploymentInfo.Metadata,
				"createdTimestamp": self.timestamp(deploymentInfo.Created),
				"updatedTimestamp": self.timestamp(deploymentInfo.Updated),
				"prepared":         deploymentInfo.Prepared,
				"approved":         deploymentInfo.Approved,
			})
			return nil
		}); err != nil {
			self.error(writer, err)
			return
		}

		self.writeJson(writer, deployments)
	} else {
		self.error(writer, err)
	}
}

func (self *Server) GetDeployment(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if deploymentInfo, err := self.Backend.GetDeployment(request.Context(), id); err == nil {
		self.writePackage(writer, deploymentInfo.Package)
	} else {
		self.error(writer, err)
	}
}
