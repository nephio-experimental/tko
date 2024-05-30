package server

import (
	"strings"

	"github.com/google/uuid"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage"
)

func IDFromListOptions(options *internalversion.ListOptions) (*string, error) {
	if (options == nil) || (options.FieldSelector == nil) {
		return nil, nil
	}

	for _, requirement := range options.FieldSelector.Requirements() {
		if (requirement.Field == "metadata.name") && isEquals(requirement.Operator) {
			if value, err := tkoutil.FromKubernetesName(requirement.Value); err == nil {
				return &value, nil
			} else {
				return nil, backendpkg.NewBadArgumentError(err.Error())
			}
		}
	}

	return nil, nil
}

func IDPatternsFromListOptions(options *internalversion.ListOptions) ([]string, error) {
	if id, err := IDFromListOptions(options); err == nil {
		if id != nil {
			return []string{*id}, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func MetadataPatternsFromListOptions(options *internalversion.ListOptions) (map[string]string, error) {
	if (options == nil) || (options.LabelSelector == nil) {
		return nil, nil
	}

	metadataPatterns := make(map[string]string)
	if requirements, ok := options.LabelSelector.Requirements(); ok {
		for _, requirement := range requirements {
			if isEquals(requirement.Operator()) {
				if values := requirement.Values(); (values != nil) && (values.Len() > 0) {
					// TODO: handle more than one value?
					value := values.UnsortedList()[0]
					if value_, err := tkoutil.FromKubernetesName(value); err == nil {
						if key, err := tkoutil.FromKubernetesName(requirement.Key()); err == nil {
							metadataPatterns[key] = value_
						} else {
							return nil, backendpkg.NewBadArgumentError(err.Error())
						}
					} else {
						return nil, backendpkg.NewBadArgumentError(err.Error())
					}
				}
			}
		}
	}
	return metadataPatterns, nil
}

var uidSpace = uuid.MustParse("b6daa158-90b7-46f7-8206-2cfce077d5d8")

func ToUID(segments ...string) types.UID {
	buffer := util.StringToBytes(strings.Join(segments, "|"))
	return types.UID(uuid.NewSHA1(uidSpace, buffer).String())
}

// Unused
func (self *Store) NewSelectionPredicate(options *internalversion.ListOptions) storage.SelectionPredicate {
	labelSelector := labels.Everything()
	fieldSelector := fields.Everything()

	if options != nil {
		if options.LabelSelector != nil {
			labelSelector = options.LabelSelector
		}
		if options.FieldSelector != nil {
			fieldSelector = options.FieldSelector
		}
	}

	return storage.SelectionPredicate{
		Label: labelSelector,
		Field: fieldSelector,
		GetAttrs: func(object runtime.Object) (labels.Set, fields.Set, error) {
			if fields, err := self.GetFieldsFunc(object); err == nil {
				return nil, fields, nil
			} else {
				return nil, nil, backendpkg.NewBadArgumentError(err.Error())
			}
		},
		//GetAttrs: storage.DefaultClusterScopedAttr,
	}
}

func isEquals(operator selection.Operator) bool {
	return (operator == selection.Equals) || (operator == selection.DoubleEquals)
}
