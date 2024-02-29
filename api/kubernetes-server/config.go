package server

import (
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	tkoopenapi "github.com/nephio-experimental/tko/api/openapi"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
)

const (
	ServerName     = "tko-apiserver"
	OpenAPITitle   = "TKO"
	OpenAPIVersion = "0.1"
)

var ConfigVersion = version.Info{
	Major: "0",
	Minor: "1",
}

func NewConfig(port int) (*apiserver.CompletedConfig, error) {
	config, err := NewRecommendedConfig(port)
	if err != nil {
		return nil, err
	}

	completedConfig := config.Complete()
	completedConfig.Version = &ConfigVersion

	return &completedConfig, nil
}

func NewRecommendedConfig(port int) (*apiserver.RecommendedConfig, error) {
	config := apiserver.NewRecommendedConfig(Codecs)

	namer := openapi.NewDefinitionNamer(Scheme)

	config.OpenAPIConfig = apiserver.DefaultOpenAPIConfig(tkoopenapi.GetOpenAPIDefinitions, namer)
	config.OpenAPIConfig.Info.Title = OpenAPITitle
	config.OpenAPIConfig.Info.Version = OpenAPIVersion

	config.OpenAPIV3Config = apiserver.DefaultOpenAPIV3Config(tkoopenapi.GetOpenAPIDefinitions, namer)
	config.OpenAPIV3Config.Info.Title = OpenAPITitle
	config.OpenAPIV3Config.Info.Version = OpenAPIVersion

	if options, err := NewRecommendedOptions(port); err == nil {
		if err := options.ApplyTo(config); err == nil {
			return config, nil
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
