package client

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ChunkSize = 100

func IsNotFoundError(err error) bool {
	if status_, ok := status.FromError(err); ok {
		if status_.Code() == codes.NotFound {
			return true
		}
	}
	return false
}

func stringifyStringList(list []string) string {
	return strings.Join(list, ",")
}

func stringifyStringMap(map_ map[string]string) string {
	var s []string
	for k, v := range map_ {
		s = append(s, k+"="+v)
	}
	return strings.Join(s, ",")
}
