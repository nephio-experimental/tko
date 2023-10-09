package backend

import (
	"regexp"
	"strings"

	"github.com/tliron/kutil/util"
)

func IdMatchesPatterns(id string, patterns []string) bool {
	if id == "" {
		return false
	}

	for _, pattern := range patterns {
		if !IdMatchesPattern(id, pattern) {
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

	for _, rune_ := range pattern {
		switch rune_ {
		case '*':
			re.WriteString(`.+`)
		default:
			re.WriteString(regexp.QuoteMeta(string(rune_)))
		}
	}

	re.WriteRune('$')
	return re.String()
}

func IdMatchesPattern(s string, pattern string) bool {
	return regexp.MustCompile(IdPatternRE(pattern)).Match(util.StringToBytes(s))
}

func IdPatternRE(pattern string) string {
	var re strings.Builder
	re.WriteRune('^')

	pattern_ := []rune(pattern)
	length := len(pattern)
	for index := 0; index < length; index++ {
		rune_ := pattern_[index]
		switch rune_ {
		case '*':
			if (index < length-1) && (pattern_[index+1] == '*') {
				// Double asterisk crosses "/" and ":" boundaries
				re.WriteString(`[0-9A-Za-z_\-\/:]+`)
				index++
			} else {
				re.WriteString(`[0-9A-Za-z_\-]+`)
			}
		default:
			re.WriteString(regexp.QuoteMeta(string(rune_)))
		}
	}

	re.WriteRune('$')
	return re.String()
}
