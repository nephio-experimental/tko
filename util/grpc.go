package util

import (
	"strings"

	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DialGRPCInsecure(address string, port int) (*grpc.ClientConn, error) {
	// See: https://github.com/grpc/grpc-go/issues/3272#issuecomment-1239710027
	address = util.JoinIPAddressPort(strings.Replace(address, "%", "%25", 1), port)

	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
