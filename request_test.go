package zttp

import (
	"bufio"
	"bytes"
	"testing"
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

	h1 := req.Get("Content-Type")
	h2 := req.Get("Content-Length")
	h3 := req.Get("Header1")
	h4 := req.Get("Header2")
	h5 := req.Get("unknown")

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
