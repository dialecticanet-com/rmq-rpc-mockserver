package comparators

import (
	"fmt"
	"regexp"
)

// Regex represents a regular expression matcher.
type Regex struct {
	Regex *regexp.Regexp
}

// NewRegex creates a new Regex matcher instance.
func NewRegex(regex string) (*Regex, error) {
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	return &Regex{
		Regex: compiled,
	}, nil
}

// Match matches the regular expression against a payload.
func (r *Regex) Match(payload []byte) bool {
	return r.Regex.Match(payload)
}
