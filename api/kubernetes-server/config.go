package server

import (
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	tkoopenapi "github.com/nephio-experimental/tko/api/openapi"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/util/feature"
	serverversion "k8s.io/apiserver/pkg/util/version"
	"k8s.io/component-base/featuregate"
	baseversion "k8s.io/component-base/version"
)

const (
	ServerName       = "tko-apiserver"
	ComponentName    = "tko"
	ComponentVersion = "0.1"
	OpenAPITitle     = "TKO"
	OpenAPIVersion   = "0.1"
)

func NewConfig(port int) (*server.CompletedConfig, error) {
	// Make sure default "kube" component is registered
	serverversion.DefaultComponentGlobalsRegistry.ComponentGlobalsOrRegister(serverversion.DefaultKubeComponent,
		serverversion.NewEffectiveVersion(baseversion.DefaultKubeBinaryVersion), feature.DefaultMutableFeatureGate)

	// Make sure our "tko" component is registered
	serverversion.DefaultComponentGlobalsRegistry.ComponentGlobalsOrRegister(
		ComponentName, serverversion.NewEffectiveVersion(ComponentVersion),
		featuregate.NewVersionedFeatureGate(version.MustParse(ComponentVersion)))

	if recommendedConfig, err := NewRecommendedConfig(port); err == nil {
		completedConfig := recommendedConfig.Complete()
		return &completedConfig, nil
	} else {
		return nil, err
	}
}

func NewRecommendedConfig(port int) (*server.RecommendedConfig, error) {
	recommendedConfig := server.NewRecommendedConfig(Codecs)

	namer := openapi.NewDefinitionNamer(Scheme)

	recommendedConfig.OpenAPIConfig = server.DefaultOpenAPIConfig(tkoopenapi.GetOpenAPIDefinitions, namer)
	recommendedConfig.OpenAPIConfig.Info.Title = OpenAPITitle
	recommendedConfig.OpenAPIConfig.Info.Version = OpenAPIVersion

	recommendedConfig.OpenAPIV3Config = server.DefaultOpenAPIV3Config(tkoopenapi.GetOpenAPIDefinitions, namer)
	recommendedConfig.OpenAPIV3Config.Info.Title = OpenAPITitle
	recommendedConfig.OpenAPIV3Config.Info.Version = OpenAPIVersion

	recommendedConfig.FeatureGate = serverversion.DefaultComponentGlobalsRegistry.FeatureGateFor(serverversion.DefaultKubeComponent)
	recommendedConfig.EffectiveVersion = serverversion.DefaultComponentGlobalsRegistry.EffectiveVersionFor(ComponentName)

	if options, err := NewRecommendedOptions(port); err == nil {
		if err := options.ApplyTo(recommendedConfig); err == nil {
			return recommendedConfig, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func NewRecommendedOptions(port int) (*options.RecommendedOptions, error) {
	options := options.NewRecommendedOptions("", Codecs.LegacyCodec(krm.SchemeGroupVersion))
	options.SecureServing.BindPort = port
	return options, nil

	/*if err := options.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{netutils.ParseIPSloppy("127.0.0.1")}); err == nil {
		return options, nil
	} else {
		return nil, err
	}*/
}
