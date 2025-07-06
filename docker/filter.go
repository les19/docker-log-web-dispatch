package docker

import (
	"strings"
)

func MatchesContainerNameFilter(containerNames []string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, name := range containerNames {
		trimmedName := strings.TrimPrefix(name, "/")
		for _, filter := range filters {
			if strings.HasPrefix(trimmedName, filter) {
				return true
			}
		}
	}
	return false
}
