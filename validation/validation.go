package validation

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/tliron/commonlog"
	validatorpkg "github.com/yannh/kubeconform/pkg/validator"
)

//
// Validation
//

type Validation struct {
	Client  *clientpkg.Client
	Timeout time.Duration
	Log     commonlog.Logger

	registeredValidators ValidatorsMap
	validators           sync.Map
	kubeconform          validatorpkg.Validator
}

func NewValidation(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger) (*Validation, error) {
	cache := filepath.Join(os.TempDir(), "tko-validation-cache")
	if err := os.MkdirAll(cache, 0700); err != nil {
		return nil, err
	}

	var kubeconform validatorpkg.Validator
	var err error
	if kubeconform, err = validatorpkg.New(nil, validatorpkg.Opts{
		Strict:               true,
		IgnoreMissingSchemas: true,
		Cache:                cache,
	}); err != nil {
		return nil, err
	}

	return &Validation{
		Client:               client,
		Timeout:              timeout,
		Log:                  log,
		registeredValidators: make(ValidatorsMap),
		kubeconform:          kubeconform,
	}, nil
}

func (self *Validation) ResetPluginCache() {
	self.validators = sync.Map{}
}
