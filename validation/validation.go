package validation

import (
	"os"
	"path/filepath"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	validatorpkg "github.com/yannh/kubeconform/pkg/validator"
)

//
// Validation
//

type Validation struct {
	Client *client.Client
	Log    commonlog.Logger

	validators  map[util.GVK]ValidatorFunc
	kubeconform validatorpkg.Validator
}

func NewValidation(client_ *client.Client, log commonlog.Logger) (*Validation, error) {
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
		Client:      client_,
		Log:         log,
		validators:  make(map[util.GVK]ValidatorFunc),
		kubeconform: kubeconform,
	}, nil
}
