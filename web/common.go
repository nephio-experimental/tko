package web

import (
	"io"
	"sort"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func sortById(info []ard.StringMap) {
	sort.Slice(info, func(i int, j int) bool {
		return info[i]["id"].(string) < info[j]["id"].(string)
	})
}

func writeResources(writer io.Writer, resources util.Resources) {
	content := make([]any, len(resources))
	for index, resource := range resources {
		content[index] = resource
	}
	(&transcribe.Transcriber{Writer: writer, Indent: "  "}).WriteYAML(content)
}
