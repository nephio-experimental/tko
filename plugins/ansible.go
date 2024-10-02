package plugins

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

const (
	Ansible = "ansible"

	AwxPrefix    = "_awx."
	AwxInventory = AwxPrefix + "inventory"
	AwxHost      = AwxPrefix + "host"
	AwxUsername  = AwxPrefix + "username"
	AwxPassword  = AwxPrefix + "password"
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
		if !strings.HasPrefix(key, AwxPrefix) {
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

func (self *AnsibleExecutor) Execute(context contextpkg.Context, input any) error {
	client := self.AWX.NewClient()

	if jobTemplateId, ok, err := client.JobTemplateIdFromName(self.JobTemplate); err == nil {
		if ok {
			// Did we already launch it?

			// TODO: this is *not* correct behavior, we are just trying to avoid running the
			// job again and again and again...

			if results, err := client.ListJobs(fmt.Sprintf("unified_job_template=%d", jobTemplateId)); err == nil {
				if count, ok := ard.With(results).Get("count").ConvertSimilar().Integer(); ok {
					if count > 0 {
						log.Infof("Ansible job already launched: %q on %q", self.JobTemplate, self.Inventory)
						return nil
					}
				} else {
					return fmt.Errorf("bad results from AWX: %v", results)
				}
			} else {
				return err
			}

			if inventoryId, ok, err := client.InventoryIdFromName(self.Inventory); err == nil {
				if ok {
					extraVars := ard.Copy(self.ExtraVars).(ard.StringMap)
					extraVars["plugin"] = input

					log.Infof("launching Ansible job: %q on %q", self.JobTemplate, self.Inventory)
					_, err := client.LaunchJob(jobTemplateId, inventoryId, extraVars)
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
	return util.NewAwxClient(self.Host, self.Username, self.Password)
}

// Utils

func newMissingPropertyError(property string) error {
	return fmt.Errorf("Ansible executor missing \"%s\" property", property)
}
