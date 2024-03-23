package client

import (
	"fmt"
	"math"
	"strings"

	api "github.com/nephio-experimental/tko/api/grpc"
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

func newWindow(offset uint, maxCount int) (*api.Window, error) {
	if offset > math.MaxUint32 {
		return nil, fmt.Errorf("offset cannot exceed %d", math.MaxUint32)
	}
	if maxCount > math.MaxInt32 {
		return nil, fmt.Errorf("maxCount cannot exceed %d", math.MaxInt32)
	}
	if maxCount < math.MinInt32 {
		return nil, fmt.Errorf("maxCount cannot be smaller than %d", math.MinInt32)
	}
	return &api.Window{
		Offset:   uint32(offset),
		MaxCount: int32(maxCount),
	}, nil
}
