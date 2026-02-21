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

// Parse parses a slice of bytes passed should be one line
// of a header
// Returns n the number of bytes read, if it completed parsing
// and a possible error
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

	key := strings.ToLower(string(data[:separatorIdx]))
	value := strings.Trim(string(line[(separatorIdx+1):]), " ")
	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid key: %s", key)
	}

	if prevVal, exists := h[key]; exists {
		h[key] = fmt.Sprintf("%s,%s", prevVal, value)
	} else {
		h[key] = value
	}

	return rn + 2, false, nil
}

func isValidKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c rune) bool {
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
		return true
	}
	return false
}
