package zttp

import (
	"fmt"
	"io"
	"net"
	"time"
)

// Mock of net.Conn struct following the net.Conn interface specifications
type MockConn struct {
	outBuf     []byte
	inBuf      []byte
	readOffset int
}

// All these functions are mocked to follow the net.Conn interface specifications
// for testing purposes only
func (m *MockConn) Read(p []byte) (n int, err error) {
	if m.readOffset >= len(m.inBuf) {
		// No more data to read
		return 0, io.EOF
	}

	n = copy(p, m.inBuf[m.readOffset:])
	m.readOffset += n
	return n, nil
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	m.outBuf = append(m.outBuf, p...)
	return len(p), nil
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Helper function to mock a request and return the response
func mockRequest(app *App, method, path, body string) string {
	conn := &MockConn{}
	req := fmt.Sprintf("%s %s HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", method, path, len(body), body)
	conn.inBuf = []byte(req)

	// Call handleClient with the mocked connection
	handleClient(conn, app)

	return string(conn.outBuf)
}
