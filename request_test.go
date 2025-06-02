package zttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
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

func TestSave(t *testing.T) {
	// Setup test directory
	testDir := filepath.Join(os.TempDir(), "zttp_save_test")

	// Cleanup after testing
	defer os.RemoveAll(testDir)

	tests := []struct {
		name        string
		formFile    *FormFile
		destination string
		setup       func() error
		expectError bool
		errorText   string
	}{
		{
			name: "Successful save",
			formFile: &FormFile{
				Filename: "test.txt",
				Content:  []byte("test content"),
			},
			destination: testDir,
			expectError: false,
		},
		{
			name:        "Nil FormFile",
			formFile:    nil,
			destination: testDir,
			expectError: true,
			errorText:   "nil FormFile",
		},
		{
			name: "Empty filename",
			formFile: &FormFile{
				Filename: "",
				Content:  []byte("content"),
			},
			destination: testDir,
			expectError: true,
			errorText:   "failed to save file",
		},
		{
			name: "Path traversal attempt",
			formFile: &FormFile{
				Filename: "./malicious.txt",
				Content:  []byte("bad content"),
			},
			destination: testDir,
			// Should sanitize successfully
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment and start fresh
			os.RemoveAll(testDir)
			os.MkdirAll(testDir, DefaultDirPerm)

			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			req := &Req{}
			err := req.Save(tt.formFile, tt.destination)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.errorText)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Verify file was saved correctly
				expectedFile := filepath.Join(tt.destination, tt.formFile.Filename)
				info, err := os.Stat(expectedFile)
				if err != nil {
					t.Fatalf("Saved file not found: %v", err)
				}

				// Verify permissions
				if info.Mode().Perm() != DefaultFilePerm {
					t.Errorf("File permissions %v != expected %v", info.Mode().Perm(), DefaultFilePerm)
				}

				// Verify content
				content, err := os.ReadFile(expectedFile)
				if err != nil {
					t.Fatalf("Failed to read saved file: %v", err)
				}
				if !bytes.Equal(content, tt.formFile.Content) {
					t.Errorf("File content mismatch")
				}
			}
		})
	}
}

func TestAccepts(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		offeredTypes []string
		expected     string
	}{
		// Exact matches
		{
			name:         "Exact match - JSON",
			acceptHeader: "application/json",
			offeredTypes: []string{"application/json", "text/html"},
			expected:     "application/json",
		},
		{
			name:         "Exact match - HTML",
			acceptHeader: "text/html",
			offeredTypes: []string{"application/json", "text/html"},
			expected:     "text/html",
		},

		// Quality values
		{
			name:         "Quality priority - prefers XML",
			acceptHeader: "text/html;q=0.8, application/xml;q=0.9",
			offeredTypes: []string{"text/html", "application/xml"},
			expected:     "application/xml",
		},
		{
			name:         "Quality fallback - lower q-value",
			acceptHeader: "text/csv;q=0.1, text/html;q=0.5",
			offeredTypes: []string{"text/csv", "text/html"},
			expected:     "text/html",
		},

		// Wildcards
		{
			name:         "Type wildcard matches any subtype",
			acceptHeader: "image/*",
			offeredTypes: []string{"image/png", "text/html"},
			expected:     "image/png",
		},
		{
			name:         "Catch-all wildcard matches first offered",
			acceptHeader: "*/*",
			offeredTypes: []string{"application/json", "text/html"},
			expected:     "application/json",
		},

		// No match cases
		{
			name:         "No overlap",
			acceptHeader: "application/xml",
			offeredTypes: []string{"text/html", "application/json"},
			expected:     "",
		},
		{
			name:         "Empty Accept header",
			acceptHeader: "",
			offeredTypes: []string{"text/html"},
			expected:     "text/html",
		},

		// Edge cases
		{
			name:         "Malformed q-value defaults to 1.0",
			acceptHeader: "text/html;q=invalid, application/json;q=0.9",
			offeredTypes: []string{"text/html", "application/json"},
			expected:     "text/html",
		},
		{
			name:         "Multiple wildcards",
			acceptHeader: "text/*, image/*",
			offeredTypes: []string{"image/png", "text/css"},
			// First in client's list with match
			expected: "text/css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Req{
				Headers: map[string]string{
					"Accept": tt.acceptHeader,
				},
			}

			result := req.Accepts(tt.offeredTypes...)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' (header: %s, offered: %v)",
					tt.expected, result, tt.acceptHeader, tt.offeredTypes)
			}
		})
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected []AcceptType
	}{
		{
			input: "text/html, application/xml;q=0.9",
			expected: []AcceptType{
				{"text/html", 1.0},
				{"application/xml", 0.9},
			},
		},
		{
			input: "text/*;q=0.5, */*;q=0.1",
			expected: []AcceptType{
				{"text/*", 0.5},
				{"*/*", 0.1},
			},
		},
		{
			// Should default to q=1.0
			input: "text/plain;q=invalid",
			expected: []AcceptType{
				{"text/plain", 1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseAcceptHeader(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d types, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i].mimeType != tt.expected[i].mimeType || result[i].q != tt.expected[i].q {
					t.Errorf("Position %d: Expected %v, got %v",
						i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		clientType  string
		offeredType string
		expected    bool
	}{
		{"text/html", "text/html", true},
		{"image/*", "image/png", true},
		{"*/*", "application/json", true},
		{"text/*", "image/png", false},
		{"application/json", "text/html", false},
	}

	for _, tt := range tests {
		t.Run(tt.clientType+"_"+tt.offeredType, func(t *testing.T) {
			result := matches(tt.clientType, tt.offeredType)
			if result != tt.expected {
				t.Errorf("Expected %t for client=%s, offered=%s",
					tt.expected, tt.clientType, tt.offeredType)
			}
		})
	}
}

func TestAcceptsCharsets(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		offered  []string
		expected string
	}{
		{
			name:     "Exact match - UTF-8",
			header:   "utf-8",
			offered:  []string{"utf-8", "iso-8859-1"},
			expected: "utf-8",
		},
		{
			name:     "Case insensitive match",
			header:   "UTF-8",
			offered:  []string{"utf-8"},
			expected: "utf-8",
		},
		{
			name:     "Quality priority",
			header:   "iso-8859-1;q=0.9, utf-8;q=0.8",
			offered:  []string{"utf-8", "iso-8859-1"},
			expected: "iso-8859-1",
		},
		{
			name:     "Wildcard acceptance",
			header:   "*",
			offered:  []string{"utf-8", "iso-8859-1"},
			expected: "utf-8",
		},
		{
			name:     "No match",
			header:   "utf-16",
			offered:  []string{"utf-8", "iso-8859-1"},
			expected: "",
		},
		{
			name:     "Empty header",
			header:   "",
			offered:  []string{"utf-8"},
			expected: "utf-8",
		},
		{
			name:     "Zero q-value ignored",
			header:   "utf-8;q=0, iso-8859-1;q=0.5",
			offered:  []string{"utf-8", "iso-8859-1"},
			expected: "iso-8859-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Req{
				Headers: map[string]string{
					"Accept-Charset": tt.header,
				},
			}
			result := req.AcceptsCharsets(tt.offered...)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' (header: %s, offered: %v)",
					tt.expected, result, tt.header, tt.offered)
			}
		})
	}
}

func TestParseAcceptCharsetHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected []AcceptCharset
	}{
		{
			input: "utf-8, iso-8859-1;q=0.8",
			expected: []AcceptCharset{
				{"utf-8", 1.0},
				{"iso-8859-1", 0.8},
			},
		},
		{
			input: "*;q=0.5, utf-16;q=0.9",
			expected: []AcceptCharset{
				{"utf-16", 0.9},
				{"*", 0.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseAcceptCharsetHeader(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d charsets, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i].charset != tt.expected[i].charset || result[i].q != tt.expected[i].q {
					t.Errorf("Position %d: Expected %v, got %v",
						i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestAcceptsEncodings(t *testing.T) {
    tests := []struct {
        name     string
        header   string
        offered  []string
        expected string
    }{
        {
            name:     "Exact match - gzip",
            header:   "gzip",
            offered:  []string{"gzip", "deflate"},
            expected: "gzip",
        },
        {
            name:     "Quality priority",
            header:   "gzip;q=0.8, deflate;q=0.9",
            offered:  []string{"gzip", "deflate"},
            expected: "deflate",
        },
        {
            name:     "Wildcard acceptance",
            header:   "*",
            offered:  []string{"br", "gzip"},
            expected: "br",
        },
        {
            name:     "Explicit identity refusal",
            header:   "gzip, identity;q=0",
            offered:  []string{"identity"},
            expected: "",
        },
        // {
        //     name:     "Implicit identity acceptance",
        //     header:   "gzip",
        //     offered:  []string{"identity"},
        //     expected: "identity",
        // },
        {
            name:     "Empty header accepts all",
            header:   "",
            offered:  []string{"gzip", "identity"},
            expected: "gzip",
        },
        {
            name:     "Case insensitive",
            header:   "GZip",
            offered:  []string{"gzip"},
            expected: "gzip",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &Req{
                Headers: map[string]string{
                    "Accept-Encoding": tt.header,
                },
            }
            result := req.AcceptsEncodings(tt.offered...)
            if result != tt.expected {
                t.Errorf("Expected '%s', got '%s' (header: %s, offered: %v)",
                    tt.expected, result, tt.header, tt.offered)
            }
        })
    }
}
