package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) timestamp(timestamp time.Time) int64 {
	return timestamp.UnixMilli()
}

func (self *Server) error(writer http.ResponseWriter, err error) {
	writer.WriteHeader(500)
	if self.Debug {
		writer.Write(util.StringToBytes(err.Error()))
	}
}

func (self *Server) writeJson(writer http.ResponseWriter, content any) {
	writer.Header().Add("Content-Type", "application/json")
	if err := transcribe.NewTranscriber().SetWriter(writer).WriteJSON(content); err != nil {
		self.error(writer, err)
	}
}

func (self *Server) writeYaml(writer http.ResponseWriter, content any) {
	writer.Header().Add("Content-Type", "application/yaml")
	if err := transcribe.NewTranscriber().SetWriter(writer).SetIndentSpaces(2).WriteYAML(content); err != nil {
		self.error(writer, err)
	}
}

func (self *Server) writePackage(writer http.ResponseWriter, package_ tkoutil.Package) {
	content := make([]any, len(package_))
	for index, resource := range package_ {
		content[index] = resource
	}
	self.writeYaml(writer, content)
}

func getWindow(request *http.Request) backend.Window {
	window := backend.Window{MaxCount: -1}

	query := request.URL.Query()

	if offset := query.Get("offset"); offset != "" {
		if offset_, err := strconv.ParseUint(offset, 10, 64); err == nil {
			window.Offset = uint(offset_)
		}
	}

	if count := query.Get("count"); count != "" {
		if count_, err := strconv.ParseInt(count, 10, 64); err == nil {
			window.MaxCount = int(count_)
		}
	}

	return window
}
