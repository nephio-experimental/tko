package server

import (
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func init() {
	runtimeutil.Must(krm.AddToScheme(Scheme))

	// See: https://github.com/kubernetes/sample-apiserver/blob/bd85aa5bdde49be0c9aa611678f08059bd2a248f/pkg/apiserver/apiserver.go#L46
	v1 := schema.GroupVersion{Version: "v1"}
	Scheme.AddUnversionedTypes(v1,
		new(meta.Status),
		new(meta.APIVersions),
		new(meta.APIGroupList),
		new(meta.APIGroup),
		new(meta.APIResourceList),
	)
	meta.AddToGroupVersion(Scheme, v1)
}
