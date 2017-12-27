package tests

import (
	"sort"

	zglob "github.com/mattn/go-zglob"
)

func Glob(pattern string) []string {
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return []string{}
	} else {
		sort.Slice(matches, func(i, j int) bool { return matches[i] < matches[j] })
		return matches
	}
}
