package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listDeployments(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if deployments, err := self.Backend.ListDeployments(context, backend.ListDeployments{}); err == nil {
		deployments_ := make([]ard.StringMap, len(deployments))
		for index, deployment := range deployments {
			deployments_[index] = ard.StringMap{
				"id":       deployment.DeploymentID,
				"template": deployment.TemplateID,
				"parent":   deployment.ParentDeploymentID,
				"site":     deployment.SiteID,
				"prepared": deployment.Prepared,
				"approved": deployment.Approved,
			}
		}
		sortById(deployments_)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(deployments_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getDeployment(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	id := request.URL.Query().Get("id")
	if deployment, err := self.Backend.GetDeployment(context, id); err == nil {
		writeResources(writer, deployment.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
