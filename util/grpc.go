package util

import (
	"strings"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DialGRPCInsecure(address string, port int) (*grpc.ClientConn, error) {
	// See: https://github.com/grpc/grpc-go/issues/3272#issuecomment-1239710027
	address = util.JoinIPAddressPort(strings.Replace(address, "%", "%25", 1), port)

	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func TriggerFromAPI(apiTrigger *api.GVK) *GVK {
	if apiTrigger != nil {
		gvk := NewGVK(apiTrigger.Group, apiTrigger.Version, apiTrigger.Kind)
		return &gvk
	}
	return nil
}

func TriggerToAPI(trigger *GVK) *api.GVK {
	if trigger != nil {
		return &api.GVK{Group: trigger.Group, Version: trigger.Version, Kind: trigger.Kind}
	}
	return nil
}

func TriggersFromAPI(apiTriggers []*api.GVK) []GVK {
	triggers := make([]GVK, len(apiTriggers))
	for index, trigger := range apiTriggers {
		triggers[index] = NewGVK(trigger.Group, trigger.Version, trigger.Kind)
	}
	return triggers
}

func TriggersToAPI(triggers []GVK) []*api.GVK {
	apiTriggers := make([]*api.GVK, len(triggers))
	for index, trigger := range triggers {
		apiTriggers[index] = &api.GVK{Group: trigger.Group, Version: trigger.Version, Kind: trigger.Kind}
	}
	return apiTriggers
}
