package utils

import (
	"fmt"
	"strings"
)

func ToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return strings.TrimSpace(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(fmt.Sprint(v), "\n", " "), "\t", " "), "\r", " "))
	}
}

func LookupPathInMap(m map[string]any, path string) (any, bool) {
	current := any(m)
	for _, p := range strings.Split(path, ".") {
		mv, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := mv[p]
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}
