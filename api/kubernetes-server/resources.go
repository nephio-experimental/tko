package server

import (
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

func PackageToKRM(package_ tkoutil.Package) *krm.Package {
	krmPackage := &krm.Package{
		Resources: make([]ard.StringMap, len(package_)),
	}
	for index, resource := range package_ {
		krmPackage.Resources[index] = ard.CopyMapsToStringMaps(resource).(ard.StringMap)
	}
	return krmPackage
}

func PackageFromKRM(krmPackage *krm.Package) tkoutil.Package {
	if krmPackage != nil {
		package_ := make(tkoutil.Package, len(krmPackage.Resources))
		for index, resource := range krmPackage.Resources {
			package_[index] = ard.CopyStringMapsToMaps(resource).(tkoutil.Resource)
		}
		return package_
	} else {
		return nil
	}
}
