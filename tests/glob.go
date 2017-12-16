package tests

import zglob "github.com/mattn/go-zglob"

func Glob(pattern string) []string {
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return []string{}
	} else {
		return matches
	}
}
