package server

import (
	"io"
	"slices"
	"strings"
	"time"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

const TimeFormat = "2006/01/02 15:04:05"

func sortById(info []ard.StringMap) {
	slices.SortFunc(info, func(a ard.StringMap, b ard.StringMap) int {
		return strings.Compare(a["id"].(string), b["id"].(string))
	})
}

func writePackage(writer io.Writer, package_ util.Package) {
	content := make([]any, len(package_))
	for index, resource := range package_ {
		content[index] = resource
	}
	transcribe.NewTranscriber().SetWriter(writer).SetIndentSpaces(2).WriteYAML(content)
}

func (self *Server) timestamp(timestamp time.Time) string {
	return timestamp.In(self.Timezone).Format(TimeFormat)
}
