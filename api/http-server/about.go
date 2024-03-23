package server

import (
	"net/http"

	"github.com/tliron/kutil/version"
)

func (self *Server) About(writer http.ResponseWriter, request *http.Request) {
	self.writeJson(writer, map[string]any{
		"instanceName":        self.InstanceName,
		"instanceDescription": self.InstanceDescription,
		"tkoVersion":          version.GitVersion,
		"backend":             self.Backend.String(),
		"http": map[string]any{
			"addressPorts": self.clientAddressPorts,
		},
	})
}
