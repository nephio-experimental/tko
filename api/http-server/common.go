package server

import (
	"io"
	"slices"
	"strings"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func sortById(info []ard.StringMap) {
	slices.SortFunc(info, func(a ard.StringMap, b ard.StringMap) int {
		return strings.Compare(a["id"].(string), b["id"].(string))
	})
}

func writeResources(writer io.Writer, resources util.Resources) {
	content := make([]any, len(resources))
	for index, resource := range resources {
		content[index] = resource
	}
	transcribe.NewTranscriber().SetWriter(writer).SetIndentSpaces(2).WriteYAML(content)
}
