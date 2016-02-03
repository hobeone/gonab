package types

import "regexp"

// RegexpUtil embed regexp.Regexp in a new type so we can extend it
type RegexpUtil struct {
	*regexp.Regexp
}

// FindStringSubmatchMap returns named matches in a map
func (r *RegexpUtil) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}
