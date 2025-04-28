package zttp

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

type Req struct {
	Method string
	Path   string
	Body   string
	Params map[string]string
}

func extractHeaders(rdr *bufio.Reader) ([]string, int) {
	headers := make([]string, 0)
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

		headers = append(headers, line)
	}

	return headers, contentLength
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
