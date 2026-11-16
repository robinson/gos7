package gos7

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"
)

// recordingTransport keeps the request and answers with a fixed response.
type recordingTransport struct {
	request  []byte
	response []byte
	err      error
}

func (r *recordingTransport) Send(request []byte) ([]byte, error) {
	r.request = append([]byte(nil), request...)
	return r.response, r.err
}

// clockResponse builds a successful Get Clock reply: result 0, return code
// 0xFF and the 10 byte S7 time field at 33 (reserved byte, century and the
// 8 byte DATE_AND_TIME at 35).
func clockResponse(value time.Time) []byte {
	response := make([]byte, 43)
	response[29] = 0xFF
	response[34] = 0x20
	var h Helper
	h.SetDateTimeAt(response, 35, value)
	return response
}

func TestPGClockWriteTransportError(t *testing.T) {
	want := errors.New("connection lost")
	client := NewClient2(&tcpPackager{}, statusTransport{err: want})
	datetime, err := client.PGClockWrite()
	if err != want || !datetime.IsZero() {
		t.Fatalf("PGClockWrite = (%v, %v), want zero time and %v", datetime, err, want)
	}
}

func TestPGClockWriteShortResponse(t *testing.T) {
	for _, length := range []int{2, 30, 31, 42} {
		response := make([]byte, length)
		if length > 29 {
			response[29] = 0xFF // success flags, but the time field is cut off
		}
		client := NewClient2(&tcpPackager{}, statusTransport{response: response})
		datetime, err := client.PGClockWrite()
		if err == nil || err.Error() != ErrorText(errIsoInvalidPDU) || !datetime.IsZero() {
			t.Fatalf("length %d: PGClockWrite = (%v, %v), want zero time and invalid PDU", length, datetime, err)
		}
	}
}

func TestPGClockWriteCPUError(t *testing.T) {
	response := make([]byte, 43)
	binary.BigEndian.PutUint16(response[27:], 0xd401)
	client := NewClient2(&tcpPackager{}, statusTransport{response: response})
	datetime, err := client.PGClockWrite()
	if err == nil || err.Error() != ErrorText(errCliInvalidPlcAnswer) || !datetime.IsZero() {
		t.Fatalf("PGClockWrite = (%v, %v), want zero time and invalid CPU answer", datetime, err)
	}
}

func TestPGClockWriteValue(t *testing.T) {
	want := time.Date(2026, time.November, 11, 18, 10, 8, 120*int(time.Millisecond), time.UTC)
	client := NewClient2(&tcpPackager{}, statusTransport{response: clockResponse(want)})
	got, err := client.PGClockWrite()
	if err != nil || !got.Equal(want) {
		t.Fatalf("PGClockWrite = (%v, %v), want (%v, nil)", got, err, want)
	}
}

func TestPGClockReadTransportError(t *testing.T) {
	want := errors.New("connection lost")
	client := NewClient2(&tcpPackager{}, &recordingTransport{err: want})
	if err := client.PGClockRead(time.Now()); err != want {
		t.Fatalf("PGClockRead error = %v, want %v", err, want)
	}
}

func TestPGClockReadShortResponse(t *testing.T) {
	for _, length := range []int{2, 30} {
		client := NewClient2(&tcpPackager{}, &recordingTransport{response: make([]byte, length)})
		err := client.PGClockRead(time.Now())
		if err == nil || err.Error() != ErrorText(errIsoInvalidPDU) {
			t.Fatalf("length %d: PGClockRead error = %v, want invalid PDU", length, err)
		}
	}
}

func TestPGClockReadCPUError(t *testing.T) {
	response := make([]byte, 31)
	binary.BigEndian.PutUint16(response[27:], 0xd401)
	client := NewClient2(&tcpPackager{}, &recordingTransport{response: response})
	err := client.PGClockRead(time.Now())
	if err == nil || err.Error() != ErrorText(errCliInvalidPlcAnswer) {
		t.Fatalf("PGClockRead error = %v, want invalid CPU answer", err)
	}
}

// TestPGClockReadRequest checks the time field of the Set Clock request: the
// 10 bytes at 29 hold a reserved byte, the century and the 8 byte
// DATE_AND_TIME, inside the 39 byte telegram.
func TestPGClockReadRequest(t *testing.T) {
	value := time.Date(2026, time.November, 11, 18, 10, 8, 0, time.UTC)
	transport := &recordingTransport{response: make([]byte, 31)}
	client := NewClient2(&tcpPackager{}, transport)
	if err := client.PGClockRead(value); err != nil {
		t.Fatal(err)
	}
	if len(transport.request) != len(s7SetDatetimeTelegram) {
		t.Fatalf("request length %d, want %d", len(transport.request), len(s7SetDatetimeTelegram))
	}
	want := make([]byte, 10)
	want[1] = 0x20
	var h Helper
	h.SetDateTimeAt(want, 2, value)
	if got := transport.request[29:39]; !bytes.Equal(got, want) {
		t.Fatalf("time field % x, want % x", got, want)
	}
}
