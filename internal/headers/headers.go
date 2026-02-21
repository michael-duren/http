package headers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/michael-duren/http/internal/parsing"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	rn := bytes.Index(data, []byte(parsing.CLRF))
	if rn == -1 {
		return 0, false, nil
	}
	if rn == 0 {
		return 0, true, nil
	}

	line := data[:rn]
	separatorIdx := bytes.IndexRune(line, ':')
	if separatorIdx == -1 {
		return 0, false, fmt.Errorf("no key value separator in header string %s", string(data))
	}

	key := string(data[:separatorIdx])
	value := strings.Trim(string(line[(separatorIdx+1):]), " ")
	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid key: %s", key)
	}

	h[key] = value
	return rn + 2, false, nil
}

func isValidKey(s string) bool {
	return len(s) > 0 && s[0] != ' ' && s[len(s)-1] != ' '
}
