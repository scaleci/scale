package tests

import (
	"os"
	"path/filepath"

	zglob "github.com/mattn/go-zglob"
)

func Glob(pattern string) []string {
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return []string{}
	} else {
		fullPaths := []string{}

		for _, m := range matches {
			if fi, err := os.Stat(m); err == nil && fi.Mode().IsRegular() {
				fullpath, err := filepath.Abs(m)
				if err == nil {
					fullPaths = append(fullPaths, fullpath)
				}
			}
		}

		return fullPaths
	}
}
