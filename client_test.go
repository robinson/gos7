package gos7

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// plcScript answers one request telegram with the telegrams to send back.
type plcScript func(request []byte) [][]byte

// newPipeClient returns a client whose TCPClientHandler is connected through
// a net.Pipe to a scripted PLC. The handler is in the state after ISO connect
// and PDU negotiation, so only job telegrams cross the pipe.
func newPipeClient(t *testing.T, timeout time.Duration, script plcScript) Client {
	t.Helper()
	clientEnd, plcEnd := net.Pipe()
	handler := &TCPClientHandler{}
	handler.conn = clientEnd
	handler.Timeout = timeout
	handler.PDULength = pduSizeRequested
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			request, err := readTelegram(plcEnd)
			if err != nil {
				return // pipe closed by the test
			}
			for _, telegram := range script(request) {
				if _, err := plcEnd.Write(telegram); err != nil {
					return
				}
			}
		}
	}()
	t.Cleanup(func() {
		clientEnd.Close()
		plcEnd.Close()
		<-done
	})
	return NewClient(handler)
}

// readTelegram reads one ISO-on-TCP telegram (TPKT length prefixed).
func readTelegram(conn net.Conn) ([]byte, error) {
	telegram := make([]byte, 4)
	if _, err := io.ReadFull(conn, telegram); err != nil {
		return nil, err
	}
	telegram = append(telegram, make([]byte, int(binary.BigEndian.Uint16(telegram[2:]))-4)...)
	_, err := io.ReadFull(conn, telegram[4:])
	return telegram, err
}

// readVarReply builds the reply to a single item Read Var job: PDU reference
// ref, return code 0xFF, byte transport size and the payload, whose length
// the item announces in bits.
func readVarReply(ref uint16, payload []byte) []byte {
	reply := []byte{
		3, 0, 0, 0, // TPKT, length set below
		2, 240, 128, // COTP DT
		s7ProtocolID, 3, 0, 0, // S7 header: protocol id, Ack-Data, redundancy id
		0, 0, // PDU reference, set below
		0, 2, // parameter length
		0, 0, // data length, set below
		0, 0, // error class and code
		4, 1, // function Read Var, one item
		0xFF, tsResByte, // return code, transport size
		0, 0, // item length in bits, set below
	}
	binary.BigEndian.PutUint16(reply[11:], ref)
	binary.BigEndian.PutUint16(reply[15:], uint16(len(payload)+4))
	binary.BigEndian.PutUint16(reply[23:], uint16(len(payload)*8))
	reply = append(reply, payload...)
	binary.BigEndian.PutUint16(reply[2:], uint16(len(reply)))
	return reply
}

// writeVarReply builds the reply to a single item Write Var job.
func writeVarReply(ref uint16, returnCode byte) []byte {
	reply := []byte{
		3, 0, 0, 22, // TPKT
		2, 240, 128, // COTP DT
		s7ProtocolID, 3, 0, 0, // S7 header: protocol id, Ack-Data, redundancy id
		0, 0, // PDU reference, set below
		0, 2, // parameter length
		0, 1, // data length
		0, 0, // error class and code
		5, 1, // function Write Var, one item
		returnCode,
	}
	binary.BigEndian.PutUint16(reply[11:], ref)
	return reply
}

// echoReply answers every request with its own PDU reference and payload.
func echoReply(payload []byte) plcScript {
	return func(request []byte) [][]byte {
		return [][]byte{readVarReply(pduReference(request), payload)}
	}
}

// pattern returns n distinct bytes starting at seed.
func pattern(n int, seed byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = seed + byte(i)
	}
	return b
}

func TestReadAreaCopiesMatchingReply(t *testing.T) {
	want := pattern(79, 1)
	client := newPipeClient(t, time.Second, echoReply(want))
	got := make([]byte, 79)
	if err := client.AGReadDB(1, 0, 79, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("AGReadDB = % x, want % x", got, want)
	}
}

func TestRequestsCarryRunningPDUReference(t *testing.T) {
	refs := make(chan uint16, 3)
	client := newPipeClient(t, time.Second, func(request []byte) [][]byte {
		refs <- pduReference(request)
		return [][]byte{readVarReply(pduReference(request), []byte{0})}
	})
	buf := make([]byte, 1)
	seen := map[uint16]bool{}
	for i := 0; i < 3; i++ {
		if err := client.AGReadDB(1, 0, 1, buf); err != nil {
			t.Fatal(err)
		}
		ref := <-refs
		if seen[ref] {
			t.Fatalf("request %d reused PDU reference 0x%04x", i, ref)
		}
		seen[ref] = true
	}
}

// TestReadAreaSkipsStaleReply replays a late answer: the reply to an earlier
// request (6 bytes for another address) arrives before the reply to the
// current one. The caller must get the current reply.
func TestReadAreaSkipsStaleReply(t *testing.T) {
	stale := pattern(6, 0x80)
	want := pattern(79, 1)
	client := newPipeClient(t, time.Second, func(request []byte) [][]byte {
		ref := pduReference(request)
		return [][]byte{readVarReply(ref-1, stale), readVarReply(ref, want)}
	})
	got := make([]byte, 79)
	if err := client.AGReadDB(1, 0, 79, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("AGReadDB = % x, want % x", got, want)
	}
}

// TestReadAreaDoesNotAcceptStaleReply: only a stale reply arrives. Before the
// fix its 6 bytes landed in the buffer, followed by 73 zero bytes, with a nil
// error. Now the read times out and the buffer is untouched.
func TestReadAreaDoesNotAcceptStaleReply(t *testing.T) {
	stale := pattern(6, 0x80)
	client := newPipeClient(t, 200*time.Millisecond, func(request []byte) [][]byte {
		return [][]byte{readVarReply(pduReference(request)-1, stale)}
	})
	sentinel := bytes.Repeat([]byte{0xAA}, 79)
	got := append([]byte(nil), sentinel...)
	err := client.AGReadDB(1, 0, 79, got)
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("AGReadDB error = %v, want deadline exceeded", err)
	}
	if !bytes.Equal(got, sentinel) {
		t.Fatalf("buffer modified by a rejected reply: % x", got)
	}
}

// TestReadAreaRejectsShortItem: the PLC echoes the reference but the item does
// not hold the requested bytes. Before the fix readArea copied past the end of
// the telegram into the caller's buffer, zero padding it silently.
func TestReadAreaRejectsShortItem(t *testing.T) {
	cases := map[string]func(ref uint16) []byte{
		"item shorter than requested": func(ref uint16) []byte {
			return readVarReply(ref, pattern(6, 1))
		},
		"announced length shorter than payload": func(ref uint16) []byte {
			reply := readVarReply(ref, pattern(79, 1))
			binary.BigEndian.PutUint16(reply[23:], 6*8)
			return reply
		},
		"announced length longer than payload": func(ref uint16) []byte {
			reply := readVarReply(ref, pattern(6, 1))
			binary.BigEndian.PutUint16(reply[23:], 79*8)
			return reply
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			client := newPipeClient(t, time.Second, func(request []byte) [][]byte {
				return [][]byte{build(pduReference(request))}
			})
			sentinel := bytes.Repeat([]byte{0xAA}, 79)
			got := append([]byte(nil), sentinel...)
			err := client.AGReadDB(1, 0, 79, got)
			if err == nil || !strings.Contains(err.Error(), ErrorText(errCliInvalidDataSizeRecvd)) {
				t.Fatalf("AGReadDB error = %v, want %s", err, ErrorText(errCliInvalidDataSizeRecvd))
			}
			if !bytes.Equal(got, sentinel) {
				t.Fatalf("buffer modified by a rejected reply: % x", got)
			}
		})
	}
}

// TestWriteAreaSkipsStaleReply: a stale Write Var acknowledgement must not
// confirm a write the PLC has not seen; the matching reply decides.
func TestWriteAreaSkipsStaleReply(t *testing.T) {
	client := newPipeClient(t, time.Second, func(request []byte) [][]byte {
		ref := pduReference(request)
		return [][]byte{writeVarReply(ref-1, 0xFF), writeVarReply(ref, code7AddressOutOfRange)}
	})
	err := client.AGWriteDB(1, 0, 2, []byte{1, 2})
	if err == nil || err.Error() != ErrorText(errCliAddressOutOfRange) {
		t.Fatalf("AGWriteDB error = %v, want %s", err, ErrorText(errCliAddressOutOfRange))
	}
}

func TestTCPPackagerVerify(t *testing.T) {
	request := append([]byte(nil), s7ReadWriteTelegram...)
	binary.BigEndian.PutUint16(request[s7PDURefOffset:], 0x1234)
	cases := []struct {
		name     string
		request  []byte
		response []byte
		wantErr  bool
	}{
		{"matching reference", request, readVarReply(0x1234, []byte{1}), false},
		{"other reference", request, readVarReply(0x1235, []byte{1}), true},
		{"non S7 request", isoConnectionRequestTelegram, readVarReply(0x0001, nil), false},
		{"short response", request, []byte{3, 0, 0, 7, 2, 240, 128}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := (&tcpPackager{}).Verify(c.request, c.response)
			if (err != nil) != c.wantErr {
				t.Fatalf("Verify error = %v, want error %v", err, c.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), ErrorText(errCliPDUReferenceMismatch)) {
				t.Fatalf("Verify error = %v, want %s", err, ErrorText(errCliPDUReferenceMismatch))
			}
		})
	}
}

// staleTransport answers with the reference of the previous request, as a
// transport without the TCP resynchronisation would after a lost reply.
type staleTransport struct{}

func (staleTransport) Send(request []byte) ([]byte, error) {
	return readVarReply(pduReference(request)-1, []byte{1}), nil
}

// TestSendVerifiesPDUReference covers transporters other than the TCP one:
// the packager rejects a reply that does not answer the request.
func TestSendVerifiesPDUReference(t *testing.T) {
	client := NewClient2(&tcpPackager{}, staleTransport{})
	_, err := client.PLCGetStatus()
	if err == nil || !strings.Contains(err.Error(), ErrorText(errCliPDUReferenceMismatch)) {
		t.Fatalf("PLCGetStatus error = %v, want %s", err, ErrorText(errCliPDUReferenceMismatch))
	}
}
