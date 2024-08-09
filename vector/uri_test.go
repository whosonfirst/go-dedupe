package vector

import (
	"fmt"
	"os"
	"testing"
)

func TestEntempifyURI(t *testing.T) {

	tests_ok := []string{
		"{tmp}test.db?memory=shared",
		"{tmp}foo.db",
	}

	tests_fail := []string{
		"test.db?memory=shared",
	}

	for _, uri := range tests_ok {

		new_uri, tmp_path, err := EntempifyURI(uri)

		if err != nil {
			t.Fatalf("Failed to entempify URI %s, %v", uri, err)
		}

		_, err = os.Stat(tmp_path)

		if err != nil {
			t.Fatalf("Failed to stat %s (%s), %v", tmp_path, uri, err)
		}

		err = os.Remove(tmp_path)

		if err != nil {
			t.Fatalf("Failed to remove %s (%s), %v", tmp_path, uri, err)
		}

		fmt.Println(uri, new_uri)
	}

	for _, uri := range tests_fail {

		_, _, err := EntempifyURI(uri)

		if err == nil {
			t.Fatalf("Expected %s to fail", uri)
		}
	}
}
