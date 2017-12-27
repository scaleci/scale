package tests

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSplitWithoutTestResultsFile(t *testing.T) {
	partition1 := []string{"test_dir/foo/test1.txt"}
	partition2 := []string{"test_dir/foo/test2.txt"}
	partition3 := []string{"test_dir/bar.txt", "test_dir/baz.txt"}

	res := Split([]string{"test_dir/foo/test1.txt",
		"test_dir/foo/test2.txt",
		"test_dir/bar.txt",
		"test_dir/baz.txt"}, int64(0), int64(3))

	if !reflect.DeepEqual(res, partition1) {
		t.Fatal(fmt.Sprintf("Expected matches to be %+v, got %+v for index 0\n",
			partition1, res))
	}

	res = Split([]string{"test_dir/foo/test1.txt",
		"test_dir/foo/test2.txt",
		"test_dir/bar.txt",
		"test_dir/baz.txt"}, int64(1), int64(3))

	if !reflect.DeepEqual(res, partition2) {
		t.Fatal(fmt.Sprintf("Expected matches to be %+v, got %+v for index 1\n",
			partition2, res))
	}

	res = Split([]string{"test_dir/foo/test1.txt",
		"test_dir/foo/test2.txt",
		"test_dir/bar.txt",
		"test_dir/baz.txt"}, int64(2), int64(3))

	if !reflect.DeepEqual(res, partition3) {
		t.Fatal(fmt.Sprintf("Expected matches to be %+v, got %+v for index 2\n",
			partition3, res))
	}
}
