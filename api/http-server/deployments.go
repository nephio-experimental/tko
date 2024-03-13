package server

import (
	"net/http"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListDeployments(writer http.ResponseWriter, request *http.Request) {
	if deploymentInfoResults, err := self.Backend.ListDeployments(request.Context(), backend.SelectDeployments{}, backend.Window{}); err == nil {
		var deployments []ard.StringMap
		if err := util.IterateResults(deploymentInfoResults, func(deploymentInfo backend.DeploymentInfo) error {
			deployments = append(deployments, ard.StringMap{
				"id":       deploymentInfo.DeploymentID,
				"template": deploymentInfo.TemplateID,
				"parent":   deploymentInfo.ParentDeploymentID,
				"site":     deploymentInfo.SiteID,
				"metadata": deploymentInfo.Metadata,
				"created":  self.timestamp(deploymentInfo.Created),
				"updated":  self.timestamp(deploymentInfo.Updated),
				"prepared": deploymentInfo.Prepared,
				"approved": deploymentInfo.Approved,
			})
			return nil
		}); err != nil {
			writer.WriteHeader(500)
			return
		}

		sortById(deployments)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(deployments)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) GetDeployment(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if deploymentInfo, err := self.Backend.GetDeployment(request.Context(), id); err == nil {
		writePackage(writer, deploymentInfo.Package)
	} else {
		writer.WriteHeader(500)
	}
}
