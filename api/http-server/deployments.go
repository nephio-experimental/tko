package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) listDeployments(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if deploymentInfoResults, err := self.Backend.ListDeployments(context, backend.ListDeployments{}); err == nil {
		var deployments []ard.StringMap
		if err := util.IterateResults(deploymentInfoResults, func(deploymentInfo backend.DeploymentInfo) error {
			deployments = append(deployments, ard.StringMap{
				"id":       deploymentInfo.DeploymentID,
				"template": deploymentInfo.TemplateID,
				"parent":   deploymentInfo.ParentDeploymentID,
				"site":     deploymentInfo.SiteID,
				"prepared": deploymentInfo.Prepared,
				"approved": deploymentInfo.Approved,
				"metadata": deploymentInfo.Metadata,
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

func (self *Server) getDeployment(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	id := request.URL.Query().Get("id")
	if deploymentInfo, err := self.Backend.GetDeployment(context, id); err == nil {
		writeResources(writer, deploymentInfo.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
