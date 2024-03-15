package validating

import (
	contextpkg "context"

	backendpkg "github.com/nephio-experimental/tko/backend"
	validationpkg "github.com/nephio-experimental/tko/validation"
)

var _ backendpkg.Backend = new(ValidatingBackend)

//
// ValidatingBackend
//

type ValidatingBackend struct {
	Backend    backendpkg.Backend
	Validation *validationpkg.Validation
}

// Wraps an existing backend with argument validation support, including
// the running of resource validation plugins.
func NewValidatingBackend(backend backendpkg.Backend, validation *validationpkg.Validation) *ValidatingBackend {
	return &ValidatingBackend{
		Backend:    backend,
		Validation: validation,
	}
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) Connect(context contextpkg.Context) error {
	return self.Backend.Connect(context)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) Release(context contextpkg.Context) error {
	return self.Backend.Release(context)
}

// ([fmt.Stringer] interface)
// ([backend.Backend] interface)
func (self *ValidatingBackend) String() string {
	return self.Backend.String()
}
