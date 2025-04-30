package zttp

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
)

type Req struct {
	Method  string
	Path    string
	Body    string
	Headers map[string]string
	Params  map[string]string
	Queries map[string]string
}

func (req *Req) Header(key string) string {
	return req.Headers[key]
}

func (req *Req) Param(key string) string {
	return req.Params[key]
}

func (req *Req) Query(key string) string {
	return req.Queries[key]
}

func (req *Req) ParseJson(target any) error {
	return json.Unmarshal([]byte(req.Body), target)
}

func extractBody(rdr *bufio.Reader, contentLength int) string {

	body := ""

	if contentLength > 0 {
		bodyBuffer := make([]byte, contentLength)
		_, err := io.ReadFull(rdr, bodyBuffer)
		if err != nil {
			log.Println("err reading body... " + err.Error())
			return ""
		}

		body = string(bodyBuffer)
	}

	return body
}

func extractHeaders(rdr *bufio.Reader) (map[string]string, int) {
	headers := make(map[string]string)
	var contentLength int = 0

	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			log.Println("err reading headers... " + err.Error())
			return nil, 0
		}

		if strings.HasPrefix(line, "Content-Length") {
			parts := strings.Split(line, ":")
			lengthStr := strings.TrimSpace(parts[1])
			contentLength, err = strconv.Atoi(lengthStr)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}

	return headers, contentLength
}

func extractQueries(raw string) map[string]string {
	queries := make(map[string]string)

	if raw == "" {
		return queries
	}

	pairs := strings.SplitSeq(raw, "&")

	for pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}
		queries[key] = value
	}

	return queries
}
