package plugins

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

const Ansible = "ansible"

const (
	AwxInventory = "_awx.inventory"
	AwxHost      = "_awx.host"
	AwxUsername  = "_awx.username"
	AwxPassword  = "_awx.password"
)

//
// AnsibleExecutor
//

type AnsibleExecutor struct {
	JobTemplate string
	Inventory   string
	ExtraVars   map[string]any

	AWX AWX
}

func NewAnsibleExecutor(arguments []string, properties map[string]string) (*AnsibleExecutor, error) {
	if len(arguments) != 1 {
		return nil, errors.New("Ansible executor must have one argument")
	}
	jobTemplate := arguments[0]

	inventory, ok := properties[AwxInventory]
	if !ok {
		return nil, newMissingPropertyError(AwxInventory)
	}

	host, ok := properties[AwxHost]
	if !ok {
		return nil, newMissingPropertyError(AwxHost)
	}

	username, ok := properties[AwxUsername]
	if !ok {
		return nil, newMissingPropertyError(AwxUsername)
	}

	password, ok := properties[AwxPassword]
	if !ok {
		return nil, newMissingPropertyError(AwxPassword)
	}

	extraVars := make(map[string]any)
	for key, value := range properties {
		if !strings.HasPrefix(key, "_awx.") {
			extraVars[key] = ard.Copy(value)
		}
	}

	return &AnsibleExecutor{
		JobTemplate: jobTemplate,
		Inventory:   inventory,
		ExtraVars:   extraVars,
		AWX: AWX{
			Host:     host,
			Username: username,
			Password: password,
		},
	}, nil
}

func (self *AnsibleExecutor) Execute(context contextpkg.Context, input any, output any) error {
	client := self.AWX.NewClient()

	if jobTemplateId, ok, err := client.JobTemplateIdFromName(self.JobTemplate); err == nil {
		if ok {
			if inventoryId, ok, err := client.InventoryIdFromName(self.Inventory); err == nil {
				if ok {
					_, err := client.LaunchJobTemplate(jobTemplateId, inventoryId, self.ExtraVars)
					return err
				} else {
					return fmt.Errorf("inventory not found: %q", self.Inventory)
				}
			} else {
				return err
			}
		} else {
			return fmt.Errorf("job template not found: %q", self.JobTemplate)
		}
	} else {
		return err
	}
}

//
// AWX
//

type AWX struct {
	Host     string
	Username string
	Password string
}

func (self *AWX) NewClient() *util.AwxClient {
	return util.NewAwxClient(self.Host, self.Username, self.Username)
}

// Utils

func newMissingPropertyError(property string) error {
	return fmt.Errorf("Ansible executor missing \"%s\" property", property)
}
