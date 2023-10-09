package server

import (
	"github.com/nephio-experimental/tko/api/backend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	// Backend errors
	if backend.IsBadArgumentError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if backend.IsNotFoundError(err) {
		return status.Error(codes.NotFound, err.Error())
	} else if backend.IsBusyError(err) {
		return status.Error(codes.Aborted, err.Error())
	} else if backend.IsTimeoutError(err) {
		return status.Error(codes.Aborted, err.Error())
	} else {
		return status.Error(codes.Internal, err.Error())
	}
}
