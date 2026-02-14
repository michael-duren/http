package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine

	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

const clrf = "\r\n"
const bufferSize = 32

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

var httpVersionRe = regexp.MustCompile(`^HTTP/\d\.\d$`)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
	}
	for req.state != requestStateDone {
		// we need to increase buffer size to handle
		// case where we have too many unread bytes
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, readErr := reader.Read(buf[readToIndex:])

		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, readErr
		}

		readToIndex += numBytesRead

		numBytesParsed, parseErr := req.parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		if readErr != nil && errors.Is(readErr, io.EOF) {
			req.state = requestStateDone
			break
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func isValidMethod(m string) bool {
	switch m {
	case HEAD, GET, OPTIONS, TRACE, PUT, DELETE, POST, PATCH, CONNECT:
		return true
	}
	return false
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(clrf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])

	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
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

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateDone
		return n, nil
	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}
