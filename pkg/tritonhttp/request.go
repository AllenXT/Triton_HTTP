package tritonhttp

import (
	"bufio"
	"fmt"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.

func parseRequestLine(line string) ([]string, error){
	fields := strings.SplitN(line, " ", 3)
	if len(fields) != 3 {
		return fields, fmt.Errorf("could not parse the request line, wrong fields %v", fields)
	}
	return fields, nil
}

func ValidateMethod(method string) error{
	if method == "GET" {
		return nil
	}else {
		return fmt.Errorf("Invalid Method: %v", method)
	}
}

func ValidateURL(url string) error{
	if !strings.HasPrefix(url, "/") {
		return fmt.Errorf("URL doesn't start with a slash")
	}
	if url == "" {
		return fmt.Errorf("missing URL")
	}
	return nil
}

func ValidateProto(proto string) error{
	if proto == "HTTP/1.1" {
		return nil
	}else {
		return fmt.Errorf("Invalid Proto: %v", proto)
	}
}

func parseHeaderLine(header string) ([]string, error){
	fields := strings.SplitN(header, ":", 2)
	if len(fields) != 2 {
		return fields, fmt.Errorf("could not parse the request line, wrong fields %v", fields)
	}
	if fields[0] == "" {
		return fields, fmt.Errorf("the key of header is empty")
	}
	if strings.Contains(fields[0], " ") {
		return fields, fmt.Errorf("Invalid format of key")
	}
	return fields, nil
}

func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	req = &Request{}
	byteR := false
	headerMap := make(map[string]string)
	hasHost := false
	req.Close = false

	// Read start line
	line, err := ReadLine(br)
	if line != "" {
		byteR = true
	}
	if err != nil {
		return nil, byteR, err
	}

	startLine, err := parseRequestLine(line)
	if err != nil {
		return nil, byteR, err
	}
	req.Method = startLine[0]
	req.URL = startLine[1]
	req.Proto = startLine[2]

	// validate the method
	if err := ValidateMethod(req.Method); err != nil {
		return nil, byteR, err
	}

	// validate the URL
	if err := ValidateURL(req.URL); err != nil {
		return nil, byteR, err
	}

	// validate the proto
	if err := ValidateProto(req.Proto); err != nil {
		return nil, byteR, err
	}

	// Read headers
	for {
		h, err :=  ReadLine(br)
		if err != nil {
			return nil, byteR, err	
		}
		// empty line means header end
		if h == "" {
			break
		}

		header, err := parseHeaderLine(h)
		if err != nil {
			return nil, byteR, err	
		}
		key := CanonicalHeaderKey(header[0])
		val := strings.TrimSpace(header[1])

		if key == "Host" {
			req.Host = val
			hasHost = true
			continue
		}

		if key == "Connection" && val == "close" {
			req.Close = true
			continue
		}

		headerMap[key] = val
	}

	req.Header = headerMap

	// Check required headers
	if !hasHost {
		return nil, byteR, fmt.Errorf("No Host Header")
	}

	return req, byteR, nil
}
