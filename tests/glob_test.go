package tests

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRegularGlob(t *testing.T) {
	matches := Glob("test_dir/**/*.txt")
	expectedMatches := []string{
		"test_dir/bar.txt",
		"test_dir/baz.txt",
		"test_dir/foo/test1.txt",
		"test_dir/foo/test2.txt",
	}
	if !reflect.DeepEqual(matches, expectedMatches) {
		t.Fatal(fmt.Sprintf("Expected matches to be %+v, got %+v\n", expectedMatches, matches))
	}
}

func TestGlobWithInvalidPattern(t *testing.T) {
	matches := Glob("test_dir\\**/*.txt")
	expectedMatches := []string{}

	if !reflect.DeepEqual(matches, expectedMatches) {
		t.Fatal(fmt.Sprintf("Expected matches to be %+v, got %+v\n", expectedMatches, matches))
	}
}
