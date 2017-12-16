package tests

import (
	"fmt"
	"testing"
)

func TestRegularGlob(t *testing.T) {
	matches := Glob("test_dir/**/*.txt")
	if matches_len := len(matches); matches_len != 3 {
		t.Fatal(fmt.Sprintf("Expected matches to be 3, got %d\n", matches_len))
	}
}

func TestGlobWithInvalidPattern(t *testing.T) {
	matches := Glob("test_dir\\**/*.txt")
	if matches_len := len(matches); matches_len != 0 {
		t.Fatal(fmt.Sprintf("Expected matches to be 0, got %d\n", matches_len))
	}
}
