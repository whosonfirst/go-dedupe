package vector

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const TMP_REPLACEMENT string = "{tmp}"

func EntempifyURI(uri string) (string, string, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return "", "", err
	}

	p := u.Path
	q := u.Query()

	if !strings.Contains(p, TMP_REPLACEMENT) {
		return "", "", nil // fmt.Errorf("URI is missing %s string", TMP_REPLACEMENT)
	}

	suffix := strings.Replace(p, TMP_REPLACEMENT, "*-", 1)

	f, err := os.CreateTemp("", suffix)

	if err != nil {
		return "", "", err
	}

	tmp_path := f.Name()

	new_uri := fmt.Sprintf("%s?%s", tmp_path, q.Encode())
	return new_uri, tmp_path, nil
}
