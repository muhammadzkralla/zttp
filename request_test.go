package zttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"
	"time"
)

// Test extracting a specific request header
func TestExtractHeader(t *testing.T) {
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": "20",
		"Header1":        "header1",
		"Header2":        "header2",
	}

	req := Req{
		Headers: headers,
	}

	h1 := req.Header("Content-Type")
	h2 := req.Header("Content-Length")
	h3 := req.Header("Header1")
	h4 := req.Header("Header2")
	h5 := req.Header("unknown")

	if h1 != "application/json" || h2 != "20" || h3 != "header1" || h4 != "header2" || h5 != "" {
		t.Errorf("Error parsing headers")
	}
}

// Test extracting all the request headers
func TestExtractHeaders(t *testing.T) {
	// Mock some headers and create a bufio reader of them
	headers := "Content-Length: 20\r\nHeader1: header1\r\nHeader2: header2\r\n\r\n"
	rdr := bufio.NewReader(bytes.NewBufferString(headers))

	extractedHeaders, extractedLen := extractHeaders(rdr)

	if extractedHeaders["Content-Length"] != "20" {
		t.Errorf("Expected header 'Content-Length: 20', but got %s", extractedHeaders["Content-Length"])
	}

	if extractedHeaders["Header1"] != "header1" {
		t.Errorf("Expected header 'Header1: header1', but got %s", extractedHeaders["Header1"])
	}

	if extractedHeaders["Header2"] != "header2" {
		t.Errorf("Expected header 'Header2: header2', but got %s", extractedHeaders["Header2"])
	}

	if extractedLen != 20 {
		t.Errorf("Expected Content-Length 20, but got %d", extractedLen)
	}
}

// Test extracting the request body
func TestExtractBody(t *testing.T) {
	// Mock a request body
	body := "Hello, world!"
	rdr := bufio.NewReader(bytes.NewBufferString(body))

	extractedBody := extractBody(rdr, 13)

	if extractedBody != "Hello, world!" {
		t.Errorf("Expected body to be 'Hello, world!', but got %s", extractedBody)
	}
}

// Test deserializing the request body to a specific target struct
func TestParseJson(t *testing.T) {
	// Mock an incoming json request
	body := `{"name":"Zkrallah","age":21}`
	req := Req{
		Body: body,
	}

	// attempt parsing the request json into the user struct object
	var user User
	err := req.ParseJson(&user)
	if err != nil {
		t.Errorf("Expected no error while parsing json, got %v", err)
	}

	if user.Name != "Zkrallah" || user.Age != 21 {
		t.Errorf("Expected no error while parsing json, got %v", err)
	}
}

// Test extracting a certain request query
func TestParseQuery(t *testing.T) {
	queries := map[string]string{
		"userId":   "2",
		"name":     "zkrallah",
		"category": "admin",
	}

	req := Req{
		Queries: queries,
	}

	q1 := req.Query("userId")
	q2 := req.Query("name")
	q3 := req.Query("category")
	q4 := req.Query("unknown")

	if q1 != "2" || q2 != "zkrallah" || q3 != "admin" || q4 != "" {
		t.Errorf("Error parsing queries")
	}
}

// Test extracting the request queries
func TestParseQueries(t *testing.T) {
	q1 := "userId=2&name=zkrallah&category=admin"
	q2 := "userId=1&category=teacher&limit=2"
	q3 := "userId=1&category="
	q4 := ""

	qs1 := extractQueries(q1)
	qs2 := extractQueries(q2)
	qs3 := extractQueries(q3)
	qs4 := extractQueries(q4)

	if len(qs1) != 3 || len(qs2) != 3 || len(qs3) != 2 || len(qs4) != 0 {
		t.Errorf("Sizes are not correct: %d", len(qs4))
	}

	if qs1["userId"] != "2" || qs1["name"] != "zkrallah" || qs1["category"] != "admin" {
		t.Errorf("Error in parsing first query")
	}

	if qs2["userId"] != "1" || qs2["category"] != "teacher" || qs2["limit"] != "2" {
		t.Errorf("Error in parsing second query")
	}

	if qs3["userId"] != "1" || qs3["category"] != "" {
		t.Errorf("Error in parsing third query")
	}
}

// Test extracting the request cookies
func TestExtractCookies(t *testing.T) {
	headers := map[string]string{
		"Header1": "header1",
		"Header2": "header2",
		"Cookie":  "sessionId=abc123; user=zkr;",
		"Header3": "header3",
		"header4": "header4",
	}

	cookies := extractCookies(headers)

	if len(cookies) != 2 {
		t.Errorf("Expected 2 cookies, but found: %d", len(cookies))
	}

	if cookies["sessionId"] != "abc123" || cookies["user"] != "zkr" {
		t.Errorf("Queries don't match expectations")
	}

	headers = map[string]string{
		"Cookie": "badcookie; valid=1; foo=bar=baz",
	}

	cookies = extractCookies(headers)

	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookies, but found: %d", len(cookies))
	}

	if cookies["valid"] != "1" {
		t.Errorf("Queries don't match expectations")
	}

}

func TestFresh(t *testing.T) {
	httpTimeFormat := "Mon, 02 Jan 2006 15:04:05 GMT"

	now := time.Now().UTC()
	lastModified := now.Format(httpTimeFormat)
	oldModified := now.Add(-24 * time.Hour).Format(httpTimeFormat)
	futureModified := now.Add(24 * time.Hour).Format(httpTimeFormat)

	tests := []struct {
		name        string
		reqHeaders  map[string]string   // request headers (single values)
		resHeaders  map[string][]string // response headers (slice values)
		expected    bool
		description string
	}{
		{
			name:       "Unconditional request",
			reqHeaders: map[string]string{},
			resHeaders: map[string][]string{
				"ETag":          {`"abc"`},
				"Last-Modified": {lastModified},
			},
			expected:    false,
			description: "Should return false when no conditional headers",
		},
		{
			name:        "ETag match",
			reqHeaders:  map[string]string{"If-None-Match": `"abc"`},
			resHeaders:  map[string][]string{"ETag": {`"abc"`}},
			expected:    true,
			description: "Should return true when ETag matches",
		},
		{
			name:        "Weak ETag match",
			reqHeaders:  map[string]string{"If-None-Match": `W/"abc"`},
			resHeaders:  map[string][]string{"ETag": {`"abc"`}},
			expected:    true,
			description: "Should handle weak ETag comparison",
		},
		{
			name:        "If-Modified-Since newer",
			reqHeaders:  map[string]string{"If-Modified-Since": futureModified},
			resHeaders:  map[string][]string{"Last-Modified": {lastModified}},
			expected:    true,
			description: "Should return true when resource not modified since",
		},
		{
			name:        "If-Modified-Since older",
			reqHeaders:  map[string]string{"If-Modified-Since": oldModified},
			resHeaders:  map[string][]string{"Last-Modified": {lastModified}},
			expected:    false,
			description: "Should return false when resource was modified",
		},
		{
			name: "No-Cache directive",
			reqHeaders: map[string]string{
				"If-None-Match": `"abc"`,
				"Cache-Control": "no-cache",
			},
			resHeaders:  map[string][]string{"ETag": {`"abc"`}},
			expected:    false,
			description: "Should bypass cache when no-cache present",
		},
		// Case 1: Only If-None-Match (matches)
		{
			name: "If-None-Match match",
			reqHeaders: map[string]string{
				"If-None-Match": "version1",
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "ETag match should return 304",
		},

		// Case 2: If-None-Match + If-Modified-Since (same date)
		{
			name: "If-None-Match with same modified date",
			reqHeaders: map[string]string{
				"If-None-Match":     "version1",
				"If-Modified-Since": lastModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "ETag match with same date should return 304",
		},

		// Case 3: If-None-Match + older If-Modified-Since
		{
			name: "If-None-Match with older date",
			reqHeaders: map[string]string{
				"If-None-Match":     "version1",
				"If-Modified-Since": oldModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "ETag match should override older modified date",
		},

		// Case 4: If-None-Match + newer If-Modified-Since
		{
			name: "If-None-Match with newer date",
			reqHeaders: map[string]string{
				"If-None-Match":     "version1",
				"If-Modified-Since": futureModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "ETag match should override newer modified date",
		},

		// Case 5: Only If-Modified-Since (same date)
		{
			name: "Only If-Modified-Since (same date)",
			reqHeaders: map[string]string{
				"If-Modified-Since": lastModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "Same modified date should return 304",
		},

		// Case 6: Only If-Modified-Since (newer date)
		{
			name: "Only If-Modified-Since (newer date)",
			reqHeaders: map[string]string{
				"If-Modified-Since": futureModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    true,
			description: "Newer modified date should return 304",
		},

		// Case 7: Only If-Modified-Since (older date)
		{
			name: "Only If-Modified-Since (older date)",
			reqHeaders: map[string]string{
				"If-Modified-Since": oldModified,
			},
			resHeaders: map[string][]string{
				"ETag":          {"version1"},
				"Last-Modified": {lastModified},
			},
			expected:    false,
			description: "Older modified date should return 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Req{
				Headers: tt.reqHeaders,
				Ctx: &Ctx{
					Req: &Req{
						Headers: tt.reqHeaders,
					},
					Res: &Res{
						Headers: tt.resHeaders,
					},
				},
			}

			if got := req.Fresh(); got != tt.expected {
				t.Errorf("%s\nFresh() = %v, want %v", tt.description, got, tt.expected)
			}
		})
	}
}

func TestCompareETags(t *testing.T) {
	tests := []struct {
		client   string
		server   string
		expected bool
	}{
		{`"abc"`, `"abc"`, true},
		{`W/"abc"`, `"abc"`, true},
		{`"abc"`, `W/"abc"`, true},
		{`W/"abc"`, `W/"abc"`, true},
		{`"abc"`, `"xyz"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.client+" vs "+tt.server, func(t *testing.T) {
			if got := compareETags(tt.client, tt.server); got != tt.expected {
				t.Errorf("compareETags(%q, %q) = %v, want %v", tt.client, tt.server, got, tt.expected)
			}
		})
	}
}

func TestIsEtagStale(t *testing.T) {
	tests := []struct {
		name        string
		etag        string
		noneMatch   string
		expected    bool
		description string
	}{
		// Exact matches
		{
			name:        "Strong match",
			etag:        `"abc"`,
			noneMatch:   `"abc"`,
			expected:    false,
			description: "Exact strong ETag match should be fresh",
		},
		{
			name:        "Weak match",
			etag:        `W/"abc"`,
			noneMatch:   `W/"abc"`,
			expected:    false,
			description: "Exact weak ETag match should be fresh",
		},

		// Weak comparison rules (RFC 7232 Section 2.3.2)
		{
			name:        "Weak vs strong match",
			etag:        `"abc"`,
			noneMatch:   `W/"abc"`,
			expected:    false,
			description: "Weak comparison should match strong ETag",
		},
		{
			name:        "Strong vs weak match",
			etag:        `W/"abc"`,
			noneMatch:   `"abc"`,
			expected:    false,
			description: "Weak ETag should match strong comparison",
		},

		// Multiple ETags
		{
			name:        "Multiple ETags with match",
			etag:        `"xyz"`,
			noneMatch:   `"abc", "xyz"`,
			expected:    false,
			description: "Should match when one of multiple ETags matches",
		},
		{
			name:        "Multiple ETags no match",
			etag:        `"123"`,
			noneMatch:   `"abc", "xyz"`,
			expected:    true,
			description: "Should be stale when no ETags match",
		},

		// Special cases
		// {
		// 	name:        "Wildcard match",
		// 	etag:        `"any"`,
		// 	noneMatch:   `*`,
		// 	expected:    false,
		// 	description: "Wildcard should match any ETag",
		// },
		{
			name:        "Empty If-None-Match",
			etag:        `"abc"`,
			noneMatch:   ``,
			expected:    true,
			description: "Empty header should be stale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEtagStale(tt.etag, []byte(tt.noneMatch))
			if result != tt.expected {
				t.Errorf("%s\nisEtagStale(%q, %q) = %v, want %v",
					tt.description, tt.etag, tt.noneMatch, result, tt.expected)
			}
		})
	}
}

func TestHasNoCacheDirective(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		// Exact matches
		{"no-cache", true, "exact match"},
		{" no-cache ", true, "with spaces"},
		{"public, no-cache", true, "in list"},
		{"no-cache, must-revalidate", true, "first in list"},

		// Invalid cases
		{"nocache", false, "missing hyphen"},
		{"no-cachex", false, "suffix characters"},
		{"xno-cache", false, "prefix characters"},
		{"no--cache", false, "double hyphen"},

		// Edge cases
		{"no-cache=", false, "with equals"},
		{"NO-CACHE", false, "case sensitive"},
		{"", false, "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := hasNoCacheDirective(tt.input); got != tt.expected {
				t.Errorf("hasNoCacheDirective(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormValue(t *testing.T) {
	// Create a test multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	writer.WriteField("username", "zkrallah")
	writer.WriteField("email", "zkrallah@zttp.com")
	writer.Close()

	tests := []struct {
		name     string
		req      *Req
		key      string
		expected string
	}{
		{
			name: "Existing field",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:      "username",
			expected: "zkrallah",
		},
		{
			name: "Non-existent field",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:      "nonexistent",
			expected: "",
		},
		{
			name: "Non-multipart request",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "{}",
			},
			key:      "username",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.FormValue(tt.key); got != tt.expected {
				t.Errorf("FormValue(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFormFile(t *testing.T) {
	// Create a test multipart form with file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a file part1
	part1, _ := writer.CreateFormFile("avatar", "test.jpg")
	io.WriteString(part1, "fake image data")

	part2, _ := writer.CreateFormFile("file", "test.pdf")
	io.WriteString(part2, "fake file data")

	writer.Close()

	tests := []struct {
		name        string
		req         *Req
		key         string
		expectError bool
		expected    *FormFile
	}{
		{
			name: "Valid image",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:         "avatar",
			expectError: false,
			expected: &FormFile{
				Filename: "test.jpg",
				Content:  []byte("fake image data"),
				Header:   textproto.MIMEHeader{},
			},
		},
		{
			name: "Valid file",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:         "file",
			expectError: false,
			expected: &FormFile{
				Filename: "test.pdf",
				Content:  []byte("fake file data"),
				Header:   textproto.MIMEHeader{},
			},
		},
		{
			name: "Non-existent file",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:         "nonexistent",
			expectError: true,
		},
		{
			name: "Non-file field",
			req: &Req{
				Headers: map[string]string{
					"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
				},
				Body: body.String(),
			},
			key:         "username",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := tt.req.FormFile(tt.key)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if file.Filename != tt.expected.Filename {
				t.Errorf("Filename = %v, want %v", file.Filename, tt.expected.Filename)
			}

			if !bytes.Equal(file.Content, tt.expected.Content) {
				t.Errorf("Content mismatch")
			}
		})
	}
}

func TestParseMultipart(t *testing.T) {
	// Helper to create multipart body
	createMultipartBody := func(boundary string) string {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		if boundary != "" {
			writer.SetBoundary(boundary)
		}
		writer.WriteField("test", "value")
		writer.Close()
		return body.String()
	}

	tests := []struct {
		name        string
		headers     map[string]string
		body        string
		expectError bool
	}{
		{
			name: "Valid multipart",
			headers: map[string]string{
				"Content-Type": "multipart/form-data; boundary=abc123",
			},
			body:        createMultipartBody("abc123"),
			expectError: false,
		},
		{
			name: "Missing boundary",
			headers: map[string]string{
				"Content-Type": "multipart/form-data",
			},
			body:        createMultipartBody(""),
			expectError: true,
		},
		{
			name: "Non-multipart content",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body:        "{}",
			expectError: true,
		},
		{
			name: "Malformed body",
			headers: map[string]string{
				"Content-Type": "multipart/form-data; boundary=abc123",
			},
			body:        "invalid multipart data",
			expectError: true,
		},
		// {
		// 	name: "Case-insensitive header",
		// 	headers: map[string]string{
		// 		"CONTENT-TYPE": "multipart/form-data; boundary=abc123",
		// 	},
		// 	body:        createMultipartBody("abc123"),
		// 	expectError: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMultipart(tt.headers, []byte(tt.body))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFormFile_LargeFile(t *testing.T) {
	// Create a large file (5MB)
	largeContent := make([]byte, 5<<20) // 5MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("largefile", "bigdata.bin")
	part.Write(largeContent)
	writer.Close()

	req := &Req{
		Headers: map[string]string{
			"Content-Type": fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()),
		},
		Body: body.String(),
	}

	file, err := req.FormFile("largefile")
	if err != nil {
		t.Fatalf("Failed to parse large file: %v", err)
	}

	if len(file.Content) != len(largeContent) {
		t.Errorf("Expected %d bytes, got %d", len(largeContent), len(file.Content))
	}
}
