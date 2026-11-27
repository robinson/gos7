package gos7

import (
	"encoding/binary"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// readTPKTFrame reads one TPKT-framed telegram (RFC 1006).
func readTPKTFrame(conn net.Conn) ([]byte, error) {
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return nil, err
	}
	n := int(binary.BigEndian.Uint16(hdr[2:]))
	if n < 4 || n > tcpMaxLength {
		return nil, io.ErrUnexpectedEOF
	}
	frame := make([]byte, n)
	copy(frame, hdr)
	if _, err := io.ReadFull(conn, frame[4:]); err != nil {
		return nil, err
	}
	return frame, nil
}

// silentNegotiationServer confirms the COTP connection request and never
// answers the S7 setup-communication request — what a PPI/Ethernet bridge
// does while its serial side hangs. It counts accepted connections and the
// connections the client has closed again.
func silentNegotiationServer(t *testing.T) (addr string, accepted, closed *atomic.Int32) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	accepted, closed = new(atomic.Int32), new(atomic.Int32)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			accepted.Add(1)
			go func(conn net.Conn) {
				defer func() { closed.Add(1); _ = conn.Close() }()
				for {
					frame, err := readTPKTFrame(conn)
					if err != nil {
						return // the client closed its end
					}
					if len(frame) == 22 && frame[5] == 0xe0 { // COTP connection request
						reply := append([]byte(nil), frame...)
						reply[5] = 0xd0 // connection confirm
						if _, err := conn.Write(reply); err != nil {
							return
						}
					}
					// everything else (the setup request) is swallowed
				}
			}(conn)
		}
	}()
	return ln.Addr().String(), accepted, closed
}

func waitCount(c *atomic.Int32, want int32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.Load() >= want {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return c.Load() >= want
}

// TestConnectClosesSocketWhenNegotiationFails: when the PDU negotiation is
// never answered, Connect fails with the negotiation error and leaves no open
// connection behind — the same as a failed ISO stage.
func TestConnectClosesSocketWhenNegotiationFails(t *testing.T) {
	addr, accepted, closed := silentNegotiationServer(t)

	h := NewTCPClientHandler(addr, 0, 2)
	h.Timeout = 200 * time.Millisecond

	err := h.Connect()
	if err == nil {
		t.Fatal("Connect must fail when the negotiation is never answered")
	}
	if !strings.Contains(err.Error(), ErrorText(errCliNegotiatingPDU)) {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted.Load() != 1 {
		t.Fatalf("server accepted %d connections, want 1", accepted.Load())
	}
	if !waitCount(closed, 1, 2*time.Second) {
		t.Fatal("the server still sees the connection open: Connect left the socket behind")
	}
	if h.conn != nil {
		t.Fatal("conn must be nil after a failed connect")
	}

	// A retry on the same handler dials again instead of reusing the dead
	// socket (tcpConnect only dials while conn == nil).
	if err := h.Connect(); err == nil {
		t.Fatal("second Connect must fail as well")
	}
	if accepted.Load() != 2 {
		t.Fatalf("second Connect must dial again, server accepted %d connections", accepted.Load())
	}
	if !waitCount(closed, 2, 2*time.Second) {
		t.Fatal("second failed Connect left its socket behind")
	}
}
