package backend

import (
	"regexp"
	"strings"

	"github.com/tliron/kutil/util"
)

func IDMatchesPatterns(id string, patterns []string) bool {
	for _, pattern := range patterns {
		if !IDMatchesPattern(id, pattern) {
			return false
		}
	}

	return true
}

func MetadataMatchesPatterns(metadata map[string]string, patterns map[string]string) bool {
	if patterns != nil {
		for key, pattern := range patterns {
			if value, ok := metadata[key]; ok {
				// TODO: different pattern for metadata
				if !MatchesPattern(value, pattern) {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}

func MatchesPattern(s string, pattern string) bool {
	return regexp.MustCompile(PatternRE(pattern)).Match(util.StringToBytes(s))
}

func PatternRE(pattern string) string {
	var re strings.Builder
	re.WriteRune('^')

	pattern_ := []rune(pattern)
	length := len(pattern_)
	for index := 0; index < length; index++ {
		rune_ := pattern_[index]
		switch rune_ {
		case '\\':
			// Escape
			if index < length-1 {
				index++
			}

		case '*':
			re.WriteString(`.*`)

		default:
			re.WriteString(regexp.QuoteMeta(string(rune_)))
		}
	}

	re.WriteRune('$')
	return re.String()
}

func EscapePatternRE(s string) string {
	return strings.ReplaceAll(s, "*", "\\*")
}

func IDMatchesPattern(s string, pattern string) bool {
	return regexp.MustCompile(IDPatternRE(pattern)).Match(util.StringToBytes(s))
}

func IDPatternRE(pattern string) string {
	var re strings.Builder
	re.WriteRune('^')

	pattern_ := []rune(pattern)
	length := len(pattern_)
	for index := 0; index < length; index++ {
		rune_ := pattern_[index]
		switch rune_ {
		case '\\':
			// Escape
			if index < length-1 {
				index++
			}

		case '*':
			if (index < length-1) && (pattern_[index+1] == '*') {
				// Double asterisk crosses "/" and ":" boundaries
				re.WriteString(`[0-9A-Za-z_.\-\/:]*`)
				index++
			} else {
				re.WriteString(`[0-9A-Za-z_.\-]*`)
			}

		default:
			re.WriteString(regexp.QuoteMeta(string(rune_)))
		}
	}

	re.WriteRune('$')
	return re.String()
}
