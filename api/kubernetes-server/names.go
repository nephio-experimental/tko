package server

import (
	"fmt"
	"strconv"
	"strings"
)

const EscapeRune rune = '$'
const MaxNameLength = 253

var EscapeString = runeToString(EscapeRune)

func IDToName(id string) (string, error) {
	id = escape(id, EscapeRune, '/', '%')
	if length := len(id); length > MaxNameLength {
		return "", fmt.Errorf("name too long: %d", length)
	}
	return id, nil
}

func NameToID(name string) (string, error) {
	var builder strings.Builder
	runes := []rune(name)
	length := len(runes)
	for index := 0; index < length; index++ {
		r := runes[index]
		if r == EscapeRune {
			if r, err := hexToRune(string(runes[index+1 : index+5])); err == nil {
				builder.WriteRune(r)
			} else {
				return "", err
			}
			index += 4
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String(), nil
}

// Utils

func escape(id string, runes ...rune) string {
	for _, r := range runes {
		id = strings.ReplaceAll(id, runeToString(r), EscapeString+fmt.Sprintf("%04x", r))
	}
	return id
}

func runeToString(r rune) string {
	return fmt.Sprintf("%c", r)
}

func runeToHex(r rune) string {
	return fmt.Sprintf("%04x", r)
}

func hexToRune(hex string) (rune, error) {
	if r, err := strconv.ParseInt(hex, 16, 32); err == nil {
		return rune(r), nil
	} else {
		return 0, err
	}
}
