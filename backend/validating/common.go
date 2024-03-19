package validating

import (
	contextpkg "context"
	"errors"
	"regexp"

	backendpkg "github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

var (
	ParallelBufferSize = 1000
	ParallelWorkers    = 10

	validIdRe = regexp.MustCompile(`^[0-9A-Za-z_.\-\/:]+$`)
)

func IsValidID(id string) bool {
	return validIdRe.MatchString(id)
}

func ValidateWindow(window *backendpkg.Window) error {
	if window.MaxCount > int(backendpkg.MaxMaxCount) {
		return backendpkg.NewBadArgumentErrorf("maxCount is too large: %d > %d", window.MaxCount, backendpkg.MaxMaxCount)
	}

	return nil
}

func ParallelDelete[R any, T any](context contextpkg.Context, results util.Results[R], getTask func(result R) T, delete_ func(task T) error) error {
	deleter := util.NewParallelExecutor[T](ParallelBufferSize, func(task T) error {
		err := delete_(task)
		// Swallow not-found errors
		if backendpkg.IsNotFoundError(err) {
			err = nil
		}
		return err
	})

	deleter.Start(ParallelWorkers)

	if err := util.IterateResults(results, func(result R) error {
		deleter.Queue(getTask(result))
		return nil
	}); err != nil {
		deleter.Close()
		return err
	}

	return errors.Join(deleter.Wait()...)
}
