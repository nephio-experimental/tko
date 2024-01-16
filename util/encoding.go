package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func EncodeResources(format string, resources Resources) ([]byte, error) {
	if resources == nil {
		resources = Resources{}
	}

	switch format {
	case "yaml":
		content := make([]any, len(resources))
		for index, resource := range resources {
			content[index] = resource
		}

		var buffer bytes.Buffer
		if err := transcribe.NewTranscriber().SetWriter(&buffer).SetIndentSpaces(2).WriteYAML(content); err == nil {
			return buffer.Bytes(), nil
		} else {
			return nil, err
		}

	case "cbor":
		return cbor.Marshal(resources)

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}

func DecodeResources(format string, content []byte) (Resources, error) {
	switch format {
	case "yaml":
		return ReadResources(format, bytes.NewReader(content))

	case "cbor":
		var resources Resources
		if err := cbor.Unmarshal(content, &resources); err == nil {
			return resources, nil
		} else {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}

func ReadResources(format string, reader io.Reader) (Resources, error) {
	switch format {
	case "yaml":
		if resources, err := ard.ReadAllYAML(reader); err == nil {
			resources_ := make(Resources, len(resources))
			var ok bool
			for index, resource := range resources {
				if resources_[index], ok = resource.(Resource); !ok {
					return nil, errors.New("a resource is not a map")
				}
				/*if _, ok := GetResourceIdentifier(resources_[index]); !ok {
					return nil, fmt.Errorf("a resource is malformed: %s", resources_[index])
				}*/
			}
			return resources_, nil
		} else {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}
