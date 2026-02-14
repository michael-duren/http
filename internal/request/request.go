package request

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var httpVersionRe = regexp.MustCompile(`^HTTP/\d\.\d$`)

func RequestFromReader(reader io.Reader) (*Request, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return nil, errors.New("unable to scan request line")
	}
	requestLineStr := scanner.Text()
	rl, err := parseRequestLine(requestLineStr)
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: *rl,
	}, nil
}

const (
	HEAD    = "HEAD"
	GET     = "GET"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE"
	PUT     = "PUT"
	DELETE  = "DELETE"
	POST    = "POST"
	PATCH   = "PATCH"
	CONNECT = "CONNECT"
)

func isValidMethod(m string) bool {
	switch m {
	case HEAD, GET, OPTIONS, TRACE, PUT, DELETE, POST, PATCH, CONNECT:
		return true
	}
	return false
}

func parseRequestLine(str string) (*RequestLine, error) {
	words := strings.Split(str, " ")
	if len(words) != 3 {
		return nil, fmt.Errorf("incorrectly formatted request line: %s", str)
	}
	method := words[0]
	if !isValidMethod(method) {
		return nil, fmt.Errorf("method specified is not a valid http verb: %s", method)
	}

	path := words[1]

	httpVersion := words[2]
	if !httpVersionRe.MatchString(httpVersion) {
		return nil, fmt.Errorf("invalid http version: %s", httpVersion)
	}
	vArr := strings.Split(httpVersion, "/")
	if len(vArr) != 2 {
		return nil, fmt.Errorf("invalid http version: %s", httpVersion)
	}
	version := vArr[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unsuported version: %s, only 1.1 is supported", version)
	}

	return &RequestLine{
		HttpVersion:   vArr[1],
		RequestTarget: path,
		Method:        method,
	}, nil
}
