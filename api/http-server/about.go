package server

import (
	"net/http"

	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/version"
)

func (self *Server) About(writer http.ResponseWriter, request *http.Request) {
	transcribe.NewTranscriber().SetWriter(writer).WriteJSON(map[string]any{
		"instanceName":        self.InstanceName,
		"instanceDescription": self.InstanceDescription,
		"tkoVersion":          version.GitVersion,
		"backend":             self.Backend.String(),
		"http": map[string]any{
			"ipStack":      string(self.IPStack),
			"addressPorts": self.clientAddressPorts,
		},
	})
}
