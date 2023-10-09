package client

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func IsNotFoundError(err error) bool {
	if status_, ok := status.FromError(err); ok {
		if status_.Code() == codes.NotFound {
			return true
		}
	}
	return false
}
