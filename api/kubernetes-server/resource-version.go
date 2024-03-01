package server

import (
	"strconv"
	"time"

	backendpkg "github.com/nephio-experimental/tko/backend"
)

func ToResourceVersion(updated time.Time) string {
	return strconv.FormatInt(updated.UnixMicro(), 10)
}

func FromResourceVersion(resourceVersion string) (time.Time, error) {
	if resourceVersion == "" {
		return time.Time{}, nil
	}

	if updated, err := strconv.ParseInt(resourceVersion, 10, 64); err == nil {
		return time.UnixMicro(updated), nil
	} else {
		return time.Time{}, backendpkg.NewBadArgumentError(err.Error())
	}
}

func ResourceVersionsEqual(a time.Time, b time.Time) bool {
	a = a.Truncate(time.Millisecond)
	b = b.Truncate(time.Millisecond)
	return a.Equal(b)
}
