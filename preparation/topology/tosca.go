package topology

import (
	contextpkg "context"
	"errors"
	"strings"

	"github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

var TOSCAGVK = util.NewGVK("topology.nephio.org", "v1alpha1", "TOSCA")

// ([preparation.PrepareFunc] signature)
func PrepareTOSCA(context contextpkg.Context, preparationContext *preparation.Context) (bool, util.Package, error) {
	if tosca, ok := preparationContext.GetTargetResource(); ok {
		if url, ok := ard.With(tosca).Get("spec", "url").ConvertSimilar().String(); ok {
			parser := util.NewTOSCAParser()
			defer parser.Release()

			if err := parser.Parse(contextpkg.TODO(), url); err == nil {
				toscaResources := make(map[string]*TOSCAResource)

				for vertextId, vertex := range cloututil.GetToscaNodeTemplates(parser.Clout, "nephio::Template") {
					toscaResources[vertextId] = NewTOSCAResource(vertex)
				}
				for vertextId, vertex := range cloututil.GetToscaNodeTemplates(parser.Clout, "nephio::Site") {
					toscaResources[vertextId] = NewTOSCAResource(vertex)
				}
				for vertextId, vertex := range cloututil.GetToscaNodeTemplates(parser.Clout, "nephio::Sites") {
					toscaResources[vertextId] = NewTOSCAResource(vertex)
				}

				if err := parser.Coerce(); err != nil {
					return false, nil, err
				}

				var package_ util.Package

				for _, toscaResource := range toscaResources {
					toscaResource.FillPropertyValues(parser.Clout)
					package_ = append(package_, toscaResource.ToPackage()...)
				}

				var placementTemplates ard.List
				for vertextId, vertex := range cloututil.GetToscaNodeTemplates(parser.Clout, "nephio::Template") {
					var sites ard.List
					for _, edge := range cloututil.GetToscaRelationships(vertex, "nephio::Host") {
						siteResource := toscaResources[edge.TargetID]
						if cloututil.IsToscaType(edge.Target.Properties, "nephio::Site") || cloututil.IsToscaType(edge.Target.Properties, "nephio::Sites") {
							sites = append(sites, siteResource.SiteName)
						}
					}

					toscaResource := toscaResources[vertextId]
					placementTemplates = append(placementTemplates, ard.Map{
						"template": toscaResource.TemplateName,
						"sites":    sites,
						"merge":    toscaResource.MergePackage,
					})
				}

				package_ = util.MergePackage(preparationContext.DeploymentPackage, package_...)

				package_ = append(package_, util.Resource{
					"apiVersion": "topology.nephio.org/v1alpha1",
					"kind":       "Placement",
					"metadata": ard.Map{
						"name": "placement",
					},
					"spec": ard.Map{
						"templates": placementTemplates,
					},
				})

				if !util.SetPreparedAnnotation(tosca, true) {
					return false, nil, errors.New("malformed TOSCA resource")
				}

				return true, package_, nil
			} else {
				return false, nil, err
			}
		}
	}

	return false, nil, nil
}

//
// TOSCAResource
//

type TOSCAResource struct {
	ID           string
	Name         string
	TemplateName string
	SiteName     string
	Properties   map[string]*TOSCAProperty
	MergePackage util.Package
}

func NewTOSCAResource(vertex *clout.Vertex) *TOSCAResource {
	self := TOSCAResource{
		ID:         vertex.ID,
		Properties: make(map[string]*TOSCAProperty),
	}
	properties_ := ard.With(vertex.Properties)
	self.Name, _ = properties_.Get("name").String()
	if properties, ok := properties_.Get("properties").StringMap(); ok {
		for name, value := range properties {
			self.NewToscaProperty(name, value)
		}
	}
	return &self
}

func (self *TOSCAResource) FillPropertyValues(clout *clout.Clout) {
	if vertex, ok := clout.Vertexes[self.ID]; ok {
		if properties, ok := ard.With(vertex.Properties).Get("properties").StringMap(); ok {
			for name, value := range properties {
				if property, ok := self.Properties[name]; ok {
					property.Value = value
				}
			}
		}
	}
}

func (self *TOSCAResource) ToPackage() util.Package {
	package_ := make(map[string]util.Resource)

	for _, property := range self.Properties {
		var resource util.Resource
		var ok bool

		var resourceName string
		if property.Name != "" {
			resourceName = property.Name
		} else {
			resourceName = self.Name + property.Suffix
		}

		key := property.GVK.String() + "/" + resourceName
		if resource, ok = package_[key]; !ok {
			resource = util.Resource{
				"apiVersion": property.GVK.APIVersion(),
				"kind":       property.GVK.Kind,
				"metadata": ard.Map{
					"name": resourceName,
				},
				"spec": make(ard.Map),
			}

			if strings.HasSuffix(property.GVK.Group, ".plugin.nephio.org") {
				if !util.SetPrepareAnnotation(resource, "Postpone") {
					panic("TODO")
				}
			}

			package_[key] = resource

			if property.GVK.Equals(TemplateGVK) {
				self.TemplateName = resourceName
			} else if property.GVK.Equals(SiteGVK) || property.GVK.Equals(SitesGVK) {
				self.SiteName = resourceName
			} else {
				self.MergePackage = append(self.MergePackage, util.Resource{
					"apiVersion": property.GVK.APIVersion(),
					"kind":       property.GVK.Kind,
					"name":       resourceName,
				})
			}
		}

		ard.With(resource["spec"]).ForceGetPath(property.Target, ".").Set(property.Value)
	}

	package__ := make(util.Package, 0, len(package_))
	for _, resource := range package_ {
		package__ = append(package__, resource)
	}
	return package__
}

//
// TOSCAProperty
//

type TOSCAProperty struct {
	Value  ard.Value
	GVK    util.GVK
	Name   string
	Suffix string
	Target string
}

func (self *TOSCAResource) NewToscaProperty(name string, value ard.Value) {
	var property TOSCAProperty
	metadata_ := ard.With(value).Get("$meta", "metadata").ConvertSimilar()
	apiVersion, _ := metadata_.Get("nephio.apiVersion").String()
	kind, _ := metadata_.Get("nephio.kind").String()
	property.GVK = util.NewGVK2(apiVersion, kind)
	property.Name, _ = metadata_.Get("nephio.name").String()
	property.Suffix, _ = metadata_.Get("nephio.suffix").String()
	property.Target, _ = metadata_.Get("nephio.target").String()
	self.Properties[name] = &property
}
