package web

import (
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listDeployments(writer http.ResponseWriter, request *http.Request) {
	if deployments, err := self.Backend.ListDeployments("", "", nil, nil, nil, nil); err == nil {
		deployments_ := make([]ard.StringMap, len(deployments))
		for index, deployment := range deployments {
			deployments_[index] = ard.StringMap{
				"id":       deployment.DeploymentID,
				"template": deployment.TemplateID,
				"parent":   deployment.ParentDeploymentID,
				"site":     deployment.SiteID,
				"prepared": deployment.Prepared,
			}
		}
		sortById(deployments_)
		(&transcribe.Transcriber{Writer: writer}).WriteJSON(deployments_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getDeployment(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if deployment, err := self.Backend.GetDeployment(id); err == nil {
		writeResources(writer, deployment.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
