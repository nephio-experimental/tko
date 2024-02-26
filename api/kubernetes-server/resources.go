package server

import (
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

func ResourcesToKRM(resources tkoutil.Resources) *krm.Package {
	krmPackage := &krm.Package{
		Resources: make([]ard.StringMap, len(resources)),
	}
	for index, resource := range resources {
		krmPackage.Resources[index] = ard.CopyMapsToStringMaps(resource).(ard.StringMap)
	}
	return krmPackage
}

func ResourcesFromKRM(krmPackage *krm.Package) tkoutil.Resources {
	if krmPackage != nil {
		resources := make(tkoutil.Resources, len(krmPackage.Resources))
		for index, resource := range krmPackage.Resources {
			resources[index] = ard.CopyStringMapsToMaps(resource).(tkoutil.Resource)
		}
		return resources
	} else {
		return nil
	}
}
