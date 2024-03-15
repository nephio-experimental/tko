package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const KubernetesNameMaxLength = 253
const KubernetesNameEscapeRune rune = '-'

var KubernetesNameAllowedRE = regexp.MustCompile(`[^0-9A-Za-z\.]`)
var KubernetesNameEscapeString = runeToString(KubernetesNameEscapeRune)

// Name rules:
//   https://kubernetes.io/docs/concepts/overview/working-with-objects/names/

// Label rules:
//   https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
//
// A valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start
// and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation
// is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')

// Converts an arbitrary string into a valid Kubernetes name, if possible.
// This works by escaping illegal characters using "-" plus the Unicode code
// in hex, left-padded with spaces to 4 characters.
//
// If the result is longer than is allows (253 characters), returns an error.
func ToKubernetesName(name string) (string, error) {
	kubernetesName := escapeName(name)
	if length := len(kubernetesName); length <= KubernetesNameMaxLength {
		return kubernetesName, nil
	} else {
		return "", fmt.Errorf("Kubernetes name too long: %d", length)
	}
}

// Converts names created by [ToKubernetesName] back to their original
// by unescaping the escape sequences.
//
// Will return an error if an escape sequence is malformed.
func FromKubernetesName(kubernetesName string) (string, error) {
	var builder strings.Builder
	runes := []rune(kubernetesName)
	length := len(runes)
	for index := 0; index < length; index++ {
		r := runes[index]
		if r == KubernetesNameEscapeRune {
			if r, err := hexToRune(runes[index+1 : index+5]); err == nil {
				builder.WriteRune(r)
			} else {
				return "", fmt.Errorf("malformed escaped Kubernetes name: %s, %w", kubernetesName, err)
			}
			index += 4
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String(), nil
}

func ToKubernetesNames(map_ map[string]string) (map[string]string, error) {
	if map_ == nil {
		return nil, nil
	}

	kubernetesMap := make(map[string]string)
	for key, value := range map_ {
		var err error
		if key, err = ToKubernetesName(key); err == nil {
			if value, err = ToKubernetesName(value); err == nil {
				kubernetesMap[key] = value
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return kubernetesMap, nil
}

func FromKubernetesNames(kubernetesMap map[string]string) (map[string]string, error) {
	if kubernetesMap == nil {
		return nil, nil
	}

	map_ := make(map[string]string)
	for key, value := range kubernetesMap {
		var err error
		if key, err = FromKubernetesName(key); err == nil {
			if value, err = FromKubernetesName(value); err == nil {
				map_[key] = value
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return map_, nil
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
