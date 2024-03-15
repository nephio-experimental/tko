package backend

import (
	"slices"
	"strings"
	"time"

	"github.com/nephio-experimental/tko/util"
)

//
// TemplateInfo
//

type TemplateInfo struct {
	TemplateID    string
	Metadata      map[string]string
	Updated       time.Time // millisecond precision
	DeploymentIDs []string
}

func (self *TemplateInfo) Clone(withDeployments bool) TemplateInfo {
	if withDeployments {
		return TemplateInfo{
			TemplateID:    self.TemplateID,
			Metadata:      util.CloneStringMap(self.Metadata),
			Updated:       self.Updated,
			DeploymentIDs: util.CloneStringList(self.DeploymentIDs),
		}
	} else {
		return TemplateInfo{
			TemplateID: self.TemplateID,
			Metadata:   util.CloneStringMap(self.Metadata),
			Updated:    self.Updated,
		}
	}
}

func (self *TemplateInfo) UpdateFromPackage(package_ util.Package) {
	updateMetadataFromPackage(self.Metadata, package_)
}

func SortTemplateInfos(templateInfos []TemplateInfo) {
	slices.SortFunc(templateInfos, func(a TemplateInfo, b TemplateInfo) int {
		return strings.Compare(a.TemplateID, b.TemplateID)
	})
}

//
// Template
//

type Template struct {
	TemplateInfo
	Package util.Package
}

func NewTemplateFromBytes(templateId string, metadata map[string]string, packageFormat string, package_ []byte) (*Template, error) {
	if package__, err := util.DecodePackage(packageFormat, package_); err == nil {
		if metadata == nil {
			metadata = make(map[string]string)
		}
		return &Template{
			TemplateInfo: TemplateInfo{
				TemplateID: templateId,
				Metadata:   metadata,
			},
			Package: package__,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Template) Clone(withDeployments bool) *Template {
	return &Template{
		TemplateInfo: self.TemplateInfo.Clone(withDeployments),
		Package:      util.ClonePackage(self.Package),
	}
}

func (self *Template) UpdateFromPackage() {
	self.TemplateInfo.UpdateFromPackage(self.Package)
}

func (self *Template) EncodePackage(format string) ([]byte, error) {
	return util.EncodePackage(format, self.Package)
}

func (self *Template) AddDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.AddToStringSet(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Template) RemoveDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.RemoveFromStringSet(self.DeploymentIDs, deploymentId)
	return ok
}

//
// SelectTemplates
//

type SelectTemplates struct {
	TemplateIDPatterns []string
	MetadataPatterns   map[string]string
}
