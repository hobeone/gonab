package types

import "regexp"

// RegexpUtil embed regexp.Regexp in a new type so we can extend it
type RegexpUtil struct {
	Regex *regexp.Regexp
}

// FindStringSubmatchMap returns named matches in a map
func (r *RegexpUtil) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.Regex.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.Regex.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}
