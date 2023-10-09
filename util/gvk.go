package util

import (
	"strings"

	"github.com/tliron/go-ard"
)

//
// GVK
//

type GVK struct {
	Group   string `json:"group" yaml:"group"`
	Version string `json:"version" yaml:"version"`
	Kind    string `json:"kind" yaml:"kind"`
}

func NewGVK(group string, version string, kind string) GVK {
	return GVK{group, version, kind}
}

func NewGVK2(apiVersion string, kind string) GVK {
	group, version := ParseApiVersion(apiVersion)
	return GVK{group, version, kind}
}

func GetGVK(resource any) (GVK, bool) {
	resource_ := ard.With(resource)
	var self GVK
	if apiVersion, ok := resource_.Get("apiVersion").String(); ok {
		self.Group, self.Version = ParseApiVersion(apiVersion)
		if self.Kind, ok = resource_.Get("kind").String(); ok {
			return self, true
		}
	}
	return self, false
}

func (self GVK) Equals(gvk GVK) bool {
	return self.Equals3(gvk.Group, gvk.Version, gvk.Kind)
}

func (self GVK) Equals3(group string, version string, kind string) bool {
	return (self.Group == group) && (self.Version == version) && (self.Kind == kind)
}

func (self GVK) Is(resource Resource) bool {
	if gvk, ok := GetGVK(resource); ok {
		return self.Equals3(gvk.Group, gvk.Version, gvk.Kind)
	}
	return false
}

func (self GVK) APIVersion() string {
	if self.Group != "" {
		return self.Group + "/" + self.Version
	} else {
		return self.Version
	}
}

// ([fmt.Stringer] interface)
func (self GVK) String() string {
	return "apiVersion: " + self.APIVersion() + ", kind: " + self.Kind
}

// Utils

func ParseApiVersion(apiVersion string) (string, string) {
	split := strings.SplitN(apiVersion, "/", 2)
	if len(split) == 2 {
		return split[0], split[1]
	} else {
		return "", apiVersion
	}
}
