package server

import (
	"fmt"
	"strconv"
	"strings"
)

const MaxNameLength = 253
const EscapeRune rune = '-'

var EscapeRuneString = runeToString(EscapeRune)
var ForbiddenRunes = []rune{EscapeRune, '/', ':', '%', '|'}

var forbiddenRuneStrings []string
var forbiddenRuneReplacements []string

func init() {
	length := len(ForbiddenRunes)
	forbiddenRuneStrings = make([]string, length)
	forbiddenRuneReplacements = make([]string, length)
	for index, r := range ForbiddenRunes {
		forbiddenRuneStrings[index] = runeToString(r)
		forbiddenRuneReplacements[index] = EscapeRuneString + runeToHex(r)
	}
}

func IDToName(id string) (string, error) {
	id = escapeName(id)
	if length := len(id); length <= MaxNameLength {
		return id, nil
	} else {
		return "", fmt.Errorf("name too long: %d", length)
	}
}

func NameToID(name string) (string, error) {
	var builder strings.Builder
	runes := []rune(name)
	length := len(runes)
	for index := 0; index < length; index++ {
		r := runes[index]
		if r == EscapeRune {
			if r, err := hexToRune(runes[index+1 : index+5]); err == nil {
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

func escapeName(name string) string {
	for index, r := range forbiddenRuneStrings {
		name = strings.ReplaceAll(name, r, forbiddenRuneReplacements[index])
	}
	return name
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
