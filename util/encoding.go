package util

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

var cbor_decode cbor.DecMode

func init() {
	// TODO: unsuccessful attempts to unmarhsal from Python SDK's client
	options := cbor.DecOptions{
		DefaultMapType: reflect.TypeFor[ard.Map](),
	}

	var err error
	cbor_decode, err = options.DecMode()
	if err != nil {
		panic(err)
	}
}

func EncodePackage(format string, package_ Package) ([]byte, error) {
	if package_ == nil {
		package_ = Package{}
	}

	switch format {
	case "yaml":
		content := make([]any, len(package_))
		for index, resource := range package_ {
			content[index] = resource
		}

		var buffer bytes.Buffer
		if err := transcribe.NewTranscriber().SetWriter(&buffer).SetIndentSpaces(2).WriteYAML(content); err == nil {
			return buffer.Bytes(), nil
		} else {
			return nil, err
		}

	case "cbor":
		return cbor.Marshal(package_)

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}

func DecodePackage(format string, content []byte) (Package, error) {
	switch format {
	case "yaml":
		return ReadPackage(format, bytes.NewReader(content))

	case "cbor":
		var package_ Package
		if err := cbor_decode.Unmarshal(content, &package_); err == nil {
			return package_, nil
		} else {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}

func ReadPackage(format string, reader io.Reader) (Package, error) {
	switch format {
	case "yaml":
		if package_, err := ard.ReadAllYAML(reader); err == nil {
			package__ := make(Package, len(package_))
			var ok bool
			for index, resource := range package_ {
				if package__[index], ok = resource.(Resource); !ok {
					return nil, fmt.Errorf("YAML resource is not a map: %+v", resource)
				}
				/*if _, ok := GetResourceIdentifier(resources_[index]); !ok {
					return nil, fmt.Errorf("a resource is malformed: %s", resources_[index])
				}*/
			}
			return package__, nil
		} else {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("format not supported: %s", format)
	}
}
