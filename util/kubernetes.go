package util

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tliron/kutil/kubernetes"
	"gopkg.in/yaml.v3"
	restpkg "k8s.io/client-go/rest"
)

const KubernetesNameMaxLength = 253
const KubernetesNameEscapeRune rune = '-'

var KubernetesNameAllowedRE = regexp.MustCompile(`[^0-9A-Za-z\.]`)
var KubernetesNameEscapeString = runeToString(KubernetesNameEscapeRune)

// Converts an arbitrary string into a valid Kubernetes name, if possible.
// This works by escaping illegal characters using "-" plus the Unicode code
// in hex, left-padded with spaces to 4 characters.
//
// If the result is longer than is allows (253 characters), returns an error.
func ToKubernetesName(name string) (string, error) {
	escapedName := escapeName(name)
	if length := len(escapedName); length <= KubernetesNameMaxLength {
		return escapedName, nil
	} else {
		return "", fmt.Errorf("Kubernetes name too long: %d", length)
	}
}

// Converts names created by [ToKubernetesName] back to their original
// by unescaping the escape sequences.
//
// Will return an error if an escape sequence is malformed.
func FromKubernetesName(escapedName string) (string, error) {
	var builder strings.Builder
	runes := []rune(escapedName)
	length := len(runes)
	for index := 0; index < length; index++ {
		r := runes[index]
		if r == KubernetesNameEscapeRune {
			if r, err := hexToRune(runes[index+1 : index+5]); err == nil {
				builder.WriteRune(r)
			} else {
				return "", fmt.Errorf("malformed escaped Kubernetes name: %s, %w", escapedName, err)
			}
			index += 4
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String(), nil
}

func ExecuteKubernetesCommand(rest restpkg.Interface, config *restpkg.Config, namespace string, podName string, containerName string, arguments []string, input any, output any) error {
	input_, err := yaml.Marshal(input)
	if err != nil {
		return err
	}

	var output_ bytes.Buffer
	var stderr bytes.Buffer

	if err := kubernetes.Exec(rest, config, namespace, podName, containerName, bytes.NewReader(input_), &output_, &stderr, false, arguments...); err == nil {
		return yaml.Unmarshal(output_.Bytes(), output)
	} else {
		return withStderr(err, stderr.String())
	}
}

// Utils

func escapeName(name string) string {
	return KubernetesNameAllowedRE.ReplaceAllStringFunc(name, func(s string) string {
		return KubernetesNameEscapeString + runeToHex([]rune(s)[0])
	})
}

func runeToString(r rune) string {
	return string(r)
	//return fmt.Sprintf("%c", r)
}

func runeToHex(r rune) string {
	s := strconv.FormatInt(int64(r), 16)
	return strings.Repeat("0", 4-len(s)) + s
	//return fmt.Sprintf("%04x", r)
}

func hexToRune(hex []rune) (rune, error) {
	if r, err := strconv.ParseInt(string(hex), 16, 32); err == nil {
		return rune(r), nil
	} else {
		return 0, err
	}
}
